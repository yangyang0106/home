package service

import (
	"cmp"
	"math"
	"slices"

	"home-decision/backend/internal/model"
	"home-decision/backend/internal/store"
)

var defaultMetrics = []model.MetricDef{
	// 指标定义决定了前端展示顺序，也决定了评分时走哪种标准化规则。
	{Key: "totalPrice", Label: "总价", Note: "越低越好", Type: "lower"},
	{Key: "commuteTime", Label: "通勤", Note: "越低越好", Type: "lower"},
	{Key: "houseAge", Label: "房龄", Note: "越低越好", Type: "lower"},
	{Key: "houseTypeScore", Label: "房屋类型", Note: "按类型映射", Type: "mapped"},
	{Key: "layoutScore", Label: "户型", Note: "主观 1-10", Type: "higher"},
	{Key: "lightScore", Label: "采光", Note: "主观 1-10", Type: "higher"},
	{Key: "noiseScore", Label: "噪音", Note: "越安静越高", Type: "higher"},
	{Key: "communityScore", Label: "小区品质", Note: "环境与界面", Type: "higher"},
	{Key: "renovationScore", Label: "装修", Note: "按装修映射", Type: "mapped"},
	{Key: "propertyScore", Label: "物业观感", Note: "主观 1-10", Type: "higher"},
	{Key: "parkingScore", Label: "停车便利", Note: "主观 1-10", Type: "higher"},
	{Key: "livingConvenience", Label: "生活便利", Note: "主观 1-10", Type: "higher"},
	{Key: "comfortScore", Label: "舒适感", Note: "主观 1-10", Type: "higher"},
	{Key: "efficiencyRate", Label: "得房率", Note: "越高越好", Type: "higher"},
}

var defaultMeta = model.Meta{
	// 这里是系统内置的默认映射，用来支撑“枚举类打分”和加减分项。
	Metrics: defaultMetrics,
	HouseTypeScores: map[string]float64{
		"次新商品房": 100,
		"商品房":   92,
		"老商品房":  78,
		"动迁房":   68,
		"回迁房":   62,
		"商住":    40,
		"公寓":    50,
	},
	RenovationScores: map[string]float64{
		"毛坯": 45,
		"简装": 70,
		"精装": 92,
	},
	BonusOptions: []model.BonusOption{
		{Key: "charger", Label: "充电桩", Score: 3},
		{Key: "smartGate", Label: "智能门禁", Score: 2},
		{Key: "clubhouse", Label: "会所/健身", Score: 3},
		{Key: "track", Label: "跑道/中庭", Score: 2},
		{Key: "kidsZone", Label: "儿童设施", Score: 2},
		{Key: "parkingSpot", Label: "固定车位", Score: 3},
		{Key: "storage", Label: "储藏空间", Score: 2},
	},
	RiskOptions: []model.RiskOption{
		{Key: "secondaryRoad", Label: "临次干道", Level: "轻度", Penalty: 5},
		{Key: "streetNoise", Label: "明显临街噪音", Level: "中度", Penalty: 10},
		{Key: "oldIssues", Label: "老破小硬伤", Level: "重度", Penalty: 20},
		{Key: "resettlement", Label: "动迁/回迁流动性风险", Level: "中度", Penalty: 10},
		{Key: "tooOld", Label: "房龄过老", Level: "中度", Penalty: 10},
		{Key: "extremeFloor", Label: "楼层极端", Level: "轻度", Penalty: 5},
		{Key: "layoutDefect", Label: "户型大硬伤", Level: "重度", Penalty: 20},
		{Key: "heavyBlock", Label: "严重采光遮挡", Level: "重度", Penalty: 20},
		{Key: "propertyComplex", Label: "产权税费复杂", Level: "重度", Penalty: 20},
		{Key: "communityWeak", Label: "小区品质明显一般", Level: "中度", Penalty: 10},
	},
}

var defaultProfiles = []model.WeightProfile{
	// 默认权重代表一套开箱即用的夫妻决策模板，用户后续可以自行改写。
	{
		Role:  "me",
		Label: "我的偏好",
		Weights: map[string]float64{
			"totalPrice": 25, "commuteTime": 22, "houseAge": 12, "houseTypeScore": 8,
			"layoutScore": 10, "lightScore": 8, "noiseScore": 5, "communityScore": 5,
			"renovationScore": 2, "propertyScore": 1, "parkingScore": 1, "livingConvenience": 1,
			"comfortScore": 6, "efficiencyRate": 1,
		},
	},
	{
		Role:  "partner",
		Label: "另一半偏好",
		Weights: map[string]float64{
			"totalPrice": 10, "commuteTime": 10, "houseAge": 5, "houseTypeScore": 5,
			"layoutScore": 20, "lightScore": 20, "noiseScore": 8, "communityScore": 10,
			"renovationScore": 7, "propertyScore": 2, "parkingScore": 1, "livingConvenience": 3,
			"comfortScore": 15, "efficiencyRate": 2,
		},
	},
}

var defaultHouses = []model.House{
	{
		ID: "house-1", HouseholdID: "demo-family", CommunityName: "春申景城", ListingName: "3号楼 1202",
		ViewDate: "2026-03-17", TotalPrice: 698, UnitPrice: 76500, Area: 91.2, HouseAge: 11,
		Floor: "18F 中层", Orientation: "南北", HouseType: "商品房", Renovation: "精装",
		CommuteTime: 42, MetroTime: 10, MonthlyFee: 580, LivingConvenience: 8, EfficiencyRate: 79,
		LightScore: 8, NoiseScore: 7, LayoutScore: 8, PropertyScore: 8, CommunityScore: 8,
		ComfortScore: 8, ParkingScore: 7, BonusSelections: []string{"charger", "parkingSpot"},
		RiskSelections: []string{"secondaryRoad"}, Notes: "厨房略小，但整体顺眼。",
	},
	{
		ID: "house-2", HouseholdID: "demo-family", CommunityName: "虹桥华苑", ListingName: "5号楼 702",
		ViewDate: "2026-03-16", TotalPrice: 658, UnitPrice: 70100, Area: 93.8, HouseAge: 18,
		Floor: "7F 低区", Orientation: "南", HouseType: "动迁房", Renovation: "简装",
		CommuteTime: 35, MetroTime: 14, MonthlyFee: 420, LivingConvenience: 9, EfficiencyRate: 73,
		LightScore: 7, NoiseScore: 5, LayoutScore: 7, PropertyScore: 6, CommunityScore: 5,
		ComfortScore: 6, ParkingScore: 6, BonusSelections: []string{"kidsZone"},
		RiskSelections: []string{"resettlement", "streetNoise", "communityWeak"}, Notes: "价格占优，但界面和流动性需要慎重。",
	},
}

type ScoringService struct {
	store store.Store
}

func NewScoringService(s store.Store) *ScoringService {
	return &ScoringService{store: s}
}

func DefaultMeta() model.Meta {
	return defaultMeta
}

func DefaultProfiles() []model.WeightProfile {
	return slices.Clone(defaultProfiles)
}

func DefaultHouses() []model.House {
	return slices.Clone(defaultHouses)
}

func (s *ScoringService) BuildDashboard(householdID string) (model.Dashboard, error) {
	profiles, err := s.store.GetWeights(householdID)
	if err != nil {
		return model.Dashboard{}, err
	}
	houses, err := s.store.ListHouses(householdID)
	if err != nil {
		return model.Dashboard{}, err
	}
	return AssembleDashboard(householdID, profiles, houses, s.store.GetMeta()), nil
}

func AssembleDashboard(householdID string, profiles []model.WeightProfile, houses []model.House, meta model.Meta) model.Dashboard {
	// Dashboard 是前端单次渲染所需的聚合结果，后端统一算完再返回，
	// 可以避免前端和后端各自维护一套评分逻辑。
	computed := computeScores(houses, profiles, meta)
	return model.Dashboard{
		HouseholdID: householdID,
		Weights:     profiles,
		Houses:      computed,
		Summary:     summarize(computed),
		Meta:        meta,
	}
}

func computeScores(houses []model.House, profiles []model.WeightProfile, meta model.Meta) []model.ComputedHouse {
	started := make([]model.House, 0, len(houses))
	for _, house := range houses {
		// 草稿或空白房源不参与比较，避免把未录完整的数据拉低整体排序。
		if isStarted(house) {
			started = append(started, house)
		}
	}

	ranges := map[string][2]float64{}
	for _, metric := range meta.Metrics {
		if metric.Type == "mapped" {
			continue
		}
		// 连续型指标先按当前家庭内的房源区间做标准化，
		// 这样“好坏”永远是基于同一批待选房源相对得出的。
		minVal, maxVal := math.MaxFloat64, -math.MaxFloat64
		found := false
		for _, house := range started {
			value := metricValue(house, metric.Key)
			if !math.IsNaN(value) {
				found = true
				minVal = math.Min(minVal, value)
				maxVal = math.Max(maxVal, value)
			}
		}
		if found {
			ranges[metric.Key] = [2]float64{minVal, maxVal}
		}
	}

	myWeights := profileWeights(profiles, "me")
	partnerWeights := profileWeights(profiles, "partner")
	result := make([]model.ComputedHouse, 0, len(started))

	for _, house := range started {
		normalized := map[string]float64{}
		for _, metric := range meta.Metrics {
			normalized[metric.Key] = normalizeMetric(house, metric, meta, ranges)
		}
		// 先分别计算双方个人分，再计算一致性与共识分，
		// 最后叠加 bonus 并扣掉风险项，完整对应产品里的公式。
		myScore := weightedScore(normalized, myWeights, meta.Metrics)
		partnerScore := weightedScore(normalized, partnerWeights, meta.Metrics)
		consistency := clamp(100-math.Abs(myScore-partnerScore), 0, 100)
		consensus := 0.45*myScore + 0.45*partnerScore + 0.10*consistency
		bonus := sumBonus(house.BonusSelections, meta.BonusOptions)
		risk := sumRisk(house.RiskSelections, meta.RiskOptions)
		finalScore := clamp(consensus+float64(bonus)-float64(risk), 0, 100)

		result = append(result, model.ComputedHouse{
			House:            house,
			MyScore:          myScore,
			PartnerScore:     partnerScore,
			ConsistencyScore: consistency,
			ConsensusScore:   consensus,
			BonusScore:       bonus,
			RiskPenalty:      risk,
			FinalScore:       finalScore,
			RiskLevel:        classifyRisk(risk),
			DecisionLabel:    classifyDecision(finalScore),
			Normalized:       normalized,
		})
	}

	slices.SortFunc(result, func(a, b model.ComputedHouse) int {
		// 默认按最终分倒序，前端直接拿来展示“当前第一”和列表排序。
		return cmp.Compare(b.FinalScore, a.FinalScore)
	})
	return result
}

func summarize(houses []model.ComputedHouse) model.Summary {
	summary := model.Summary{Count: len(houses)}
	if len(houses) == 0 {
		return summary
	}
	// 汇总信息主要给首页总览卡片和推荐区使用。
	summary.BestFinalScore = houses[0].FinalScore
	total := 0.0
	for _, house := range houses {
		total += house.ConsensusScore
	}
	summary.AverageConsensus = total / float64(len(houses))
	top := houses[0]
	summary.TopChoice = &top
	if len(houses) > 1 {
		runnerUp := houses[1]
		summary.RunnerUp = &runnerUp
	}
	return summary
}

func profileWeights(profiles []model.WeightProfile, role string) map[string]float64 {
	for _, profile := range profiles {
		if profile.Role == role {
			return profile.Weights
		}
	}
	return map[string]float64{}
}

func metricValue(house model.House, key string) float64 {
	switch key {
	case "totalPrice":
		return house.TotalPrice
	case "commuteTime":
		return house.CommuteTime
	case "houseAge":
		return house.HouseAge
	case "layoutScore":
		return house.LayoutScore
	case "lightScore":
		return house.LightScore
	case "noiseScore":
		return house.NoiseScore
	case "communityScore":
		return house.CommunityScore
	case "propertyScore":
		return house.PropertyScore
	case "parkingScore":
		return house.ParkingScore
	case "livingConvenience":
		return house.LivingConvenience
	case "comfortScore":
		return house.ComfortScore
	case "efficiencyRate":
		return house.EfficiencyRate
	default:
		return math.NaN()
	}
}

func normalizeMetric(house model.House, metric model.MetricDef, meta model.Meta, ranges map[string][2]float64) float64 {
	if metric.Key == "houseTypeScore" {
		return meta.HouseTypeScores[house.HouseType]
	}
	if metric.Key == "renovationScore" {
		return meta.RenovationScores[house.Renovation]
	}

	value := metricValue(house, metric.Key)
	rangeSet, ok := ranges[metric.Key]
	if !ok || rangeSet[0] == rangeSet[1] {
		if metric.Type == "lower" {
			return clamp(100-value, 0, 100)
		}
		return clamp(value*10, 0, 100)
	}
	minVal, maxVal := rangeSet[0], rangeSet[1]
	if metric.Type == "lower" {
		return clamp(100*(maxVal-value)/(maxVal-minVal), 0, 100)
	}
	return clamp(100*(value-minVal)/(maxVal-minVal), 0, 100)
}

func weightedScore(normalized, weights map[string]float64, metrics []model.MetricDef) float64 {
	totalWeight := 0.0
	for _, value := range weights {
		totalWeight += value
	}
	if totalWeight == 0 {
		totalWeight = 1
	}
	sum := 0.0
	for _, metric := range metrics {
		sum += normalized[metric.Key] * (weights[metric.Key] / totalWeight)
	}
	return sum
}

func sumBonus(selected []string, options []model.BonusOption) int {
	total := 0
	for _, key := range selected {
		for _, option := range options {
			if option.Key == key {
				total += option.Score
			}
		}
	}
	return total
}

func sumRisk(selected []string, options []model.RiskOption) int {
	total := 0
	for _, key := range selected {
		for _, option := range options {
			if option.Key == key {
				total += option.Penalty
			}
		}
	}
	return total
}

func classifyRisk(penalty int) string {
	switch {
	case penalty == 0:
		return "无风险"
	case penalty >= 20:
		return "高风险"
	case penalty >= 10:
		return "中风险"
	default:
		return "低风险"
	}
}

func classifyDecision(score float64) string {
	switch {
	case score >= 85:
		return "强推荐"
	case score >= 75:
		return "可重点复看"
	case score >= 65:
		return "可谈可比"
	default:
		return "谨慎/放弃"
	}
}

func clamp(value, min, max float64) float64 {
	return math.Min(math.Max(value, min), max)
}

func isStarted(house model.House) bool {
	return house.CommunityName != "" || house.ListingName != "" || house.TotalPrice > 0 || house.UnitPrice > 0 || house.Area > 0 || house.Notes != ""
}
