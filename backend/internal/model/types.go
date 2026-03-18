package model

type MetricDef struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Note  string `json:"note"`
	Type  string `json:"type"`
}

type WeightProfile struct {
	Role    string             `json:"role"`
	Label   string             `json:"label"`
	Weights map[string]float64 `json:"weights"`
}

type BonusOption struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Score int    `json:"score"`
}

type RiskOption struct {
	Key     string `json:"key"`
	Label   string `json:"label"`
	Level   string `json:"level"`
	Penalty int    `json:"penalty"`
}

type House struct {
	ID                string   `json:"id"`
	HouseholdID       string   `json:"householdId"`
	CommunityName     string   `json:"communityName"`
	ListingName       string   `json:"listingName"`
	ViewDate          string   `json:"viewDate"`
	TotalPrice        float64  `json:"totalPrice"`
	UnitPrice         float64  `json:"unitPrice"`
	Area              float64  `json:"area"`
	HouseAge          float64  `json:"houseAge"`
	Floor             string   `json:"floor"`
	Orientation       string   `json:"orientation"`
	HouseType         string   `json:"houseType"`
	Renovation        string   `json:"renovation"`
	CommuteTime       float64  `json:"commuteTime"`
	MetroTime         float64  `json:"metroTime"`
	MonthlyFee        float64  `json:"monthlyFee"`
	LivingConvenience float64  `json:"livingConvenience"`
	EfficiencyRate    float64  `json:"efficiencyRate"`
	LightScore        float64  `json:"lightScore"`
	NoiseScore        float64  `json:"noiseScore"`
	LayoutScore       float64  `json:"layoutScore"`
	PropertyScore     float64  `json:"propertyScore"`
	CommunityScore    float64  `json:"communityScore"`
	ComfortScore      float64  `json:"comfortScore"`
	ParkingScore      float64  `json:"parkingScore"`
	BonusSelections   []string `json:"bonusSelections"`
	RiskSelections    []string `json:"riskSelections"`
	Notes             string   `json:"notes"`
}

type ComputedHouse struct {
	House
	MyScore          float64            `json:"myScore"`
	PartnerScore     float64            `json:"partnerScore"`
	ConsistencyScore float64            `json:"consistencyScore"`
	ConsensusScore   float64            `json:"consensusScore"`
	BonusScore       int                `json:"bonusScore"`
	RiskPenalty      int                `json:"riskPenalty"`
	FinalScore       float64            `json:"finalScore"`
	RiskLevel        string             `json:"riskLevel"`
	DecisionLabel    string             `json:"decisionLabel"`
	Normalized       map[string]float64 `json:"normalized"`
}

type Summary struct {
	Count            int            `json:"count"`
	BestFinalScore   float64        `json:"bestFinalScore"`
	AverageConsensus float64        `json:"averageConsensus"`
	TopChoice        *ComputedHouse `json:"topChoice,omitempty"`
	RunnerUp         *ComputedHouse `json:"runnerUp,omitempty"`
}

type Dashboard struct {
	HouseholdID string          `json:"householdId"`
	Weights     []WeightProfile `json:"weights"`
	Houses      []ComputedHouse `json:"houses"`
	Summary     Summary         `json:"summary"`
	Meta        Meta            `json:"meta"`
}

type Meta struct {
	Metrics          []MetricDef        `json:"metrics"`
	HouseTypeScores  map[string]float64 `json:"houseTypeScores"`
	RenovationScores map[string]float64 `json:"renovationScores"`
	BonusOptions     []BonusOption      `json:"bonusOptions"`
	RiskOptions      []RiskOption       `json:"riskOptions"`
}

type User struct {
	ID          string `json:"id"`
	LoginID     string `json:"loginId"`
	DisplayName string `json:"displayName"`
	LinkCode    string `json:"linkCode"`
	IsAdmin     bool   `json:"isAdmin"`
}

type Session struct {
	Token     string
	UserID    string
	ExpiresAt string
}

type AccountMember struct {
	UserID      string `json:"userId"`
	LoginID     string `json:"loginId"`
	DisplayName string `json:"displayName"`
	LinkCode    string `json:"linkCode"`
}

type AccountProfile struct {
	User        User            `json:"user"`
	HouseholdID string          `json:"householdId"`
	Members     []AccountMember `json:"members"`
}

type AdminUser struct {
	User
	HouseholdID string `json:"householdId"`
	CreatedAt   string `json:"createdAt"`
}
