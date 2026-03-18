const STORAGE_KEY = "home-decision-calculator-v1";

const METRIC_DEFS = [
  { key: "totalPrice", label: "总价", note: "越低越好", type: "lower" },
  { key: "commuteTime", label: "通勤", note: "越低越好", type: "lower" },
  { key: "houseAge", label: "房龄", note: "越低越好", type: "lower" },
  { key: "houseTypeScore", label: "房屋类型", note: "按类型映射", type: "mapped" },
  { key: "layoutScore", label: "户型", note: "主观 1-10", type: "higher" },
  { key: "lightScore", label: "采光", note: "主观 1-10", type: "higher" },
  { key: "noiseScore", label: "噪音", note: "越安静越高", type: "higher" },
  { key: "communityScore", label: "小区品质", note: "环境与界面", type: "higher" },
  { key: "renovationScore", label: "装修", note: "按装修映射", type: "mapped" },
  { key: "propertyScore", label: "物业观感", note: "主观 1-10", type: "higher" },
  { key: "parkingScore", label: "停车便利", note: "主观 1-10", type: "higher" },
  { key: "livingConvenience", label: "生活便利", note: "主观 1-10", type: "higher" },
  { key: "comfortScore", label: "舒适感", note: "主观 1-10", type: "higher" },
  { key: "efficiencyRate", label: "得房率", note: "越高越好", type: "higher" }
];

const DEFAULT_WEIGHTS = {
  me: {
    totalPrice: 25,
    commuteTime: 22,
    houseAge: 12,
    houseTypeScore: 8,
    layoutScore: 10,
    lightScore: 8,
    noiseScore: 5,
    communityScore: 5,
    renovationScore: 2,
    propertyScore: 1,
    parkingScore: 1,
    livingConvenience: 1,
    comfortScore: 6,
    efficiencyRate: 1
  },
  partner: {
    totalPrice: 10,
    commuteTime: 10,
    houseAge: 5,
    houseTypeScore: 5,
    layoutScore: 20,
    lightScore: 20,
    noiseScore: 8,
    communityScore: 10,
    renovationScore: 7,
    propertyScore: 2,
    parkingScore: 1,
    livingConvenience: 3,
    comfortScore: 15,
    efficiencyRate: 2
  }
};

const HOUSE_TYPE_SCORES = {
  "次新商品房": 100,
  商品房: 92,
  老商品房: 78,
  动迁房: 68,
  回迁房: 62,
  商住: 40,
  公寓: 50
};

const RENOVATION_SCORES = {
  毛坯: 45,
  简装: 70,
  精装: 92
};

const BONUS_OPTIONS = [
  { key: "charger", label: "充电桩", score: 3 },
  { key: "smartGate", label: "智能门禁", score: 2 },
  { key: "clubhouse", label: "会所/健身", score: 3 },
  { key: "track", label: "跑道/中庭", score: 2 },
  { key: "kidsZone", label: "儿童设施", score: 2 },
  { key: "parkingSpot", label: "固定车位", score: 3 },
  { key: "storage", label: "储藏空间", score: 2 }
];

const RISK_OPTIONS = [
  { key: "secondaryRoad", label: "临次干道", level: "轻度", penalty: 5 },
  { key: "streetNoise", label: "明显临街噪音", level: "中度", penalty: 10 },
  { key: "oldIssues", label: "老破小硬伤", level: "重度", penalty: 20 },
  { key: "resettlement", label: "动迁/回迁流动性风险", level: "中度", penalty: 10 },
  { key: "tooOld", label: "房龄过老", level: "中度", penalty: 10 },
  { key: "extremeFloor", label: "楼层极端", level: "轻度", penalty: 5 },
  { key: "layoutDefect", label: "户型大硬伤", level: "重度", penalty: 20 },
  { key: "heavyBlock", label: "严重采光遮挡", level: "重度", penalty: 20 },
  { key: "propertyComplex", label: "产权税费复杂", level: "重度", penalty: 20 },
  { key: "communityWeak", label: "小区品质明显一般", level: "中度", penalty: 10 }
];

const FORM_GROUPS = [
  {
    title: "基础信息",
    fields: [
      { key: "communityName", label: "小区名", type: "text", placeholder: "例如：春申景城" },
      { key: "listingName", label: "房源名/楼栋室", type: "text", placeholder: "例如：3号楼 1202" },
      { key: "viewDate", label: "看房日期", type: "date" },
      { key: "totalPrice", label: "总价", type: "number", suffix: "万", min: 0, step: 1 },
      { key: "unitPrice", label: "单价", type: "number", suffix: "元/平", min: 0, step: 100 },
      { key: "area", label: "建筑面积", type: "number", suffix: "平", min: 0, step: 0.1 },
      { key: "houseAge", label: "房龄", type: "number", suffix: "年", min: 0, step: 1 },
      { key: "floor", label: "楼层", type: "text", placeholder: "例如：中楼层 / 18F 中层" },
      { key: "orientation", label: "朝向", type: "text", placeholder: "例如：南北通透" },
      {
        key: "houseType",
        label: "房屋类型",
        type: "select",
        options: Object.keys(HOUSE_TYPE_SCORES)
      },
      {
        key: "renovation",
        label: "装修情况",
        type: "select",
        options: Object.keys(RENOVATION_SCORES)
      }
    ]
  },
  {
    title: "通勤与生活",
    fields: [
      { key: "commuteTime", label: "到公司耗时", type: "number", suffix: "分钟", min: 0, step: 1 },
      { key: "metroTime", label: "到地铁耗时", type: "number", suffix: "分钟", min: 0, step: 1 },
      { key: "monthlyFee", label: "物业费", type: "number", suffix: "元/月", min: 0, step: 10 },
      {
        key: "livingConvenience",
        label: "周边生活便利度",
        type: "range",
        min: 1,
        max: 10,
        step: 1
      },
      { key: "efficiencyRate", label: "得房率", type: "number", suffix: "%", min: 0, max: 100, step: 1 }
    ]
  },
  {
    title: "居住体验",
    fields: [
      { key: "lightScore", label: "采光", type: "range", min: 1, max: 10, step: 1 },
      { key: "noiseScore", label: "噪音", type: "range", min: 1, max: 10, step: 1 },
      { key: "layoutScore", label: "户型合理性", type: "range", min: 1, max: 10, step: 1 },
      { key: "propertyScore", label: "物业观感", type: "range", min: 1, max: 10, step: 1 },
      { key: "communityScore", label: "小区环境", type: "range", min: 1, max: 10, step: 1 },
      { key: "comfortScore", label: "室内舒适感", type: "range", min: 1, max: 10, step: 1 },
      { key: "parkingScore", label: "停车便利度", type: "range", min: 1, max: 10, step: 1 }
    ]
  }
];

const DEFAULT_HOUSE = () => ({
  id: crypto.randomUUID(),
  communityName: "",
  listingName: "",
  viewDate: new Date().toISOString().slice(0, 10),
  totalPrice: 0,
  unitPrice: 0,
  area: 0,
  houseAge: 0,
  floor: "",
  orientation: "",
  houseType: "商品房",
  renovation: "简装",
  commuteTime: 0,
  metroTime: 0,
  monthlyFee: 0,
  livingConvenience: 6,
  efficiencyRate: 75,
  lightScore: 6,
  noiseScore: 6,
  layoutScore: 6,
  propertyScore: 6,
  communityScore: 6,
  comfortScore: 6,
  parkingScore: 6,
  bonusSelections: [],
  riskSelections: [],
  notes: ""
});

const DEMO_HOUSES = [
  {
    ...DEFAULT_HOUSE(),
    communityName: "春申景城",
    listingName: "3号楼 1202",
    totalPrice: 698,
    unitPrice: 76500,
    area: 91.2,
    houseAge: 11,
    floor: "18F 中层",
    orientation: "南北",
    houseType: "商品房",
    renovation: "精装",
    commuteTime: 42,
    metroTime: 10,
    monthlyFee: 580,
    livingConvenience: 8,
    efficiencyRate: 79,
    lightScore: 8,
    noiseScore: 7,
    layoutScore: 8,
    propertyScore: 8,
    communityScore: 8,
    comfortScore: 8,
    parkingScore: 7,
    bonusSelections: ["charger", "parkingSpot"],
    riskSelections: ["secondaryRoad"],
    notes: "厨房略小，但整体顺眼。"
  },
  {
    ...DEFAULT_HOUSE(),
    communityName: "虹桥华苑",
    listingName: "5号楼 702",
    totalPrice: 658,
    unitPrice: 70100,
    area: 93.8,
    houseAge: 18,
    floor: "7F 低区",
    orientation: "南",
    houseType: "动迁房",
    renovation: "简装",
    commuteTime: 35,
    metroTime: 14,
    monthlyFee: 420,
    livingConvenience: 9,
    efficiencyRate: 73,
    lightScore: 7,
    noiseScore: 5,
    layoutScore: 7,
    propertyScore: 6,
    communityScore: 5,
    comfortScore: 6,
    parkingScore: 6,
    bonusSelections: ["kidsZone"],
    riskSelections: ["resettlement", "streetNoise", "communityWeak"],
    notes: "价格占优，但界面和流动性需要慎重。"
  }
];

let state = loadState();
let activeWeightProfile = "me";
let editingHouseId = state.houses[0]?.id ?? null;

const summaryCardsEl = document.getElementById("summary-cards");
const topPicksEl = document.getElementById("top-picks");
const weightTabsEl = document.getElementById("weight-tabs");
const weightEditorEl = document.getElementById("weight-editor");
const houseFormEl = document.getElementById("house-form");
const houseListEl = document.getElementById("house-list");
const weightRowTemplate = document.getElementById("weight-row-template");

document.getElementById("seed-demo").addEventListener("click", () => {
  state.houses = DEMO_HOUSES.map((house) => ({ ...house, id: crypto.randomUUID() }));
  editingHouseId = state.houses[0]?.id ?? null;
  persist();
  render();
});

document.getElementById("reset-weights").addEventListener("click", () => {
  state.weights = structuredClone(DEFAULT_WEIGHTS);
  persist();
  renderWeightEditor();
  renderScores();
  renderHouseList();
});

document.getElementById("new-house").addEventListener("click", () => {
  const house = DEFAULT_HOUSE();
  state.houses.unshift(house);
  editingHouseId = house.id;
  persist();
  render();
  houseFormEl.scrollIntoView({ behavior: "smooth", block: "start" });
});

function loadState() {
  const fallback = {
    weights: structuredClone(DEFAULT_WEIGHTS),
    houses: []
  };

  try {
    const saved = JSON.parse(localStorage.getItem(STORAGE_KEY));
    if (!saved) return fallback;
    return {
      weights: {
        me: { ...structuredClone(DEFAULT_WEIGHTS).me, ...(saved.weights?.me ?? {}) },
        partner: { ...structuredClone(DEFAULT_WEIGHTS).partner, ...(saved.weights?.partner ?? {}) }
      },
      houses: Array.isArray(saved.houses) ? saved.houses : []
    };
  } catch {
    return fallback;
  }
}

function persist() {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(state));
}

function render() {
  renderWeightTabs();
  renderWeightEditor();
  renderHouseForm();
  renderScores();
  renderHouseList();
}

function renderWeightTabs() {
  weightTabsEl.innerHTML = "";
  [
    { key: "me", label: "我的偏好", desc: "更看重预算与通勤" },
    { key: "partner", label: "另一半偏好", desc: "更看重采光与居住感" }
  ].forEach((item) => {
    const button = document.createElement("button");
    button.type = "button";
    button.className = `weight-tab${activeWeightProfile === item.key ? " active" : ""}`;
    button.innerHTML = `<strong>${item.label}</strong><div class="hint-text">${item.desc}</div>`;
    button.addEventListener("click", () => {
      activeWeightProfile = item.key;
      renderWeightTabs();
      renderWeightEditor();
    });
    weightTabsEl.appendChild(button);
  });
}

function renderWeightEditor() {
  weightEditorEl.innerHTML = "";
  METRIC_DEFS.forEach((metric) => {
    const row = weightRowTemplate.content.firstElementChild.cloneNode(true);
    const input = row.querySelector("input");
    const output = row.querySelector("output");
    row.querySelector(".metric-name").textContent = metric.label;
    row.querySelector(".metric-note").textContent = metric.note;
    input.value = state.weights[activeWeightProfile][metric.key];
    output.value = `${input.value}%`;
    input.addEventListener("input", () => {
      state.weights[activeWeightProfile][metric.key] = Number(input.value);
      output.value = `${input.value}%`;
      persist();
      renderScores();
      renderHouseList();
    });
    weightEditorEl.appendChild(row);
  });
}

function renderHouseForm() {
  const house = getEditingHouse();
  if (!house) return;

  houseFormEl.innerHTML = "";
  FORM_GROUPS.forEach((group, index) => {
    const section = document.createElement("section");
    section.className = "form-group";
    const fieldsGrid = document.createElement("div");
    fieldsGrid.className = `fields-grid${index < 2 ? " two-col" : ""}`;

    const title = document.createElement("h3");
    title.textContent = group.title;
    section.appendChild(title);

    group.fields.forEach((field) => {
      fieldsGrid.appendChild(buildField(field, house));
    });

    section.appendChild(fieldsGrid);
    houseFormEl.appendChild(section);
  });

  houseFormEl.appendChild(buildChoiceGroup("附加加分项", BONUS_OPTIONS, house.bonusSelections, "bonusSelections", "选中后自动累计 bonus_score"));
  houseFormEl.appendChild(buildChoiceGroup("风险扣分项", RISK_OPTIONS, house.riskSelections, "riskSelections", "风险不进主分数，只做额外罚分"));

  const notesGroup = document.createElement("section");
  notesGroup.className = "form-group";
  notesGroup.innerHTML = `
    <h3>补充备注</h3>
    <div class="field">
      <label for="notes">现场感受</label>
      <textarea id="notes" rows="4" placeholder="例如：客厅采光很好，但主卧压抑感明显。">${house.notes ?? ""}</textarea>
    </div>
  `;
  notesGroup.querySelector("textarea").addEventListener("input", (event) => {
    house.notes = event.target.value;
    persist();
    renderHouseList();
  });
  houseFormEl.appendChild(notesGroup);

  const actions = document.createElement("div");
  actions.className = "form-actions";
  const saveButton = document.createElement("button");
  saveButton.type = "button";
  saveButton.className = "primary-button";
  saveButton.textContent = "保存当前房源";
  saveButton.addEventListener("click", () => {
    persist();
    render();
  });

  const deleteButton = document.createElement("button");
  deleteButton.type = "button";
  deleteButton.className = "secondary-button";
  deleteButton.textContent = "删除当前房源";
  deleteButton.addEventListener("click", () => {
    state.houses = state.houses.filter((item) => item.id !== house.id);
    editingHouseId = state.houses[0]?.id ?? null;
    if (!editingHouseId) {
      const next = DEFAULT_HOUSE();
      state.houses.push(next);
      editingHouseId = next.id;
    }
    persist();
    render();
  });

  actions.append(saveButton, deleteButton);
  houseFormEl.appendChild(actions);
}

function buildField(field, house) {
  const wrap = document.createElement("div");
  wrap.className = field.type === "range" ? "slider-field" : "field";

  const id = `field-${field.key}`;

  if (field.type === "select") {
    wrap.innerHTML = `
      <label for="${id}">${field.label}</label>
      <select id="${id}"></select>
    `;
    const select = wrap.querySelector("select");
    field.options.forEach((option) => {
      const node = document.createElement("option");
      node.value = option;
      node.textContent = option;
      select.appendChild(node);
    });
    select.value = house[field.key];
    select.addEventListener("change", (event) => {
      house[field.key] = event.target.value;
      persist();
      renderScores();
      renderHouseList();
    });
    return wrap;
  }

  if (field.type === "range") {
    wrap.innerHTML = `
      <div class="slider-head">
        <label for="${id}">${field.label}</label>
        <output>${house[field.key]}</output>
      </div>
      <input id="${id}" type="range" min="${field.min}" max="${field.max}" step="${field.step}" value="${house[field.key]}" />
    `;
    const input = wrap.querySelector("input");
    const output = wrap.querySelector("output");
    input.addEventListener("input", (event) => {
      house[field.key] = Number(event.target.value);
      output.value = event.target.value;
      persist();
      renderScores();
      renderHouseList();
    });
    return wrap;
  }

  const inputType = field.type === "number" || field.type === "date" ? field.type : "text";
  wrap.innerHTML = `
    <label for="${id}">${field.label}</label>
    <input
      id="${id}"
      type="${inputType}"
      ${field.min !== undefined ? `min="${field.min}"` : ""}
      ${field.max !== undefined ? `max="${field.max}"` : ""}
      ${field.step !== undefined ? `step="${field.step}"` : ""}
      ${field.placeholder ? `placeholder="${field.placeholder}"` : ""}
      value="${house[field.key] ?? ""}"
    />
    ${field.suffix ? `<span class="field-help">${field.suffix}</span>` : ""}
  `;
  wrap.querySelector("input").addEventListener("input", (event) => {
    house[field.key] = field.type === "number" ? Number(event.target.value) : event.target.value;
    persist();
    renderScores();
    renderHouseList();
  });
  return wrap;
}

function buildChoiceGroup(title, options, selectedValues, stateKey, helpText) {
  const section = document.createElement("section");
  section.className = "form-group";
  section.innerHTML = `<h3>${title}</h3><div class="field-help">${helpText}</div>`;
  const list = document.createElement("div");
  list.className = "choice-grid";
  const house = getEditingHouse();

  options.forEach((option) => {
    const id = `${stateKey}-${option.key}`;
    const line = document.createElement("label");
    line.className = "choice-item";
    line.innerHTML = `
      <div>
        <div>${option.label}</div>
        <small>${option.score ? `加分 +${option.score}` : `${option.level}风险 -${option.penalty}`}</small>
      </div>
      <input id="${id}" type="checkbox" ${selectedValues.includes(option.key) ? "checked" : ""} />
    `;
    line.querySelector("input").addEventListener("change", (event) => {
      const next = new Set(house[stateKey]);
      if (event.target.checked) {
        next.add(option.key);
      } else {
        next.delete(option.key);
      }
      house[stateKey] = Array.from(next);
      persist();
      renderScores();
      renderHouseList();
    });
    list.appendChild(line);
  });

  section.appendChild(list);
  return section;
}

function renderScores() {
  const computed = computeAllHouseScores(state.houses, state.weights);
  summaryCardsEl.innerHTML = "";
  topPicksEl.innerHTML = "";

  const cards = [
    { label: "已记录房源", value: computed.length, suffix: "套" },
    {
      label: "当前最佳",
      value: computed[0] ? `${round1(computed[0].finalScore)}` : "--",
      suffix: "分"
    },
    {
      label: "平均共识分",
      value: computed.length ? round1(avg(computed.map((item) => item.consensusScore))) : "--",
      suffix: "分"
    }
  ];

  cards.forEach((card) => {
    const node = document.createElement("article");
    node.className = "summary-card";
    node.innerHTML = `<span>${card.label}</span><strong>${card.value}</strong><span>${card.suffix}</span>`;
    summaryCardsEl.appendChild(node);
  });

  if (computed.length === 0) {
    topPicksEl.innerHTML = `
      <div class="empty-state">
        <h3>还没有房源记录</h3>
        <p>先新增一套房，系统就会自动生成双人分数、共识分和风险等级。</p>
      </div>
    `;
    return;
  }

  const top = computed[0];
  const runnerUp = computed[1];

  topPicksEl.appendChild(buildTopPick("当前第一", top, "good"));
  if (runnerUp) {
    topPicksEl.appendChild(buildTopPick("备选第二", runnerUp, "warn"));
  }
}

function buildTopPick(title, item, tone) {
  const node = document.createElement("article");
  node.className = "top-pick";
  node.innerHTML = `
    <div class="top-pick__head">
      <div>
        <span>${title}</span>
        <h3>${item.communityName || "未命名房源"}</h3>
      </div>
      <span class="pill ${tone}">${item.decisionLabel}</span>
    </div>
    <p class="house-meta">${item.listingName || "待补充楼栋室"} · 风险等级：${item.riskLevel}</p>
    <div class="score-grid">
      <div class="stat-chip"><span>我的分</span><strong>${round1(item.myScore)}</strong></div>
      <div class="stat-chip"><span>她的分</span><strong>${round1(item.partnerScore)}</strong></div>
      <div class="stat-chip"><span>共识分</span><strong>${round1(item.consensusScore)}</strong></div>
      <div class="stat-chip"><span>最终分</span><strong>${round1(item.finalScore)}</strong></div>
    </div>
  `;
  return node;
}

function renderHouseList() {
  const computed = computeAllHouseScores(state.houses, state.weights);
  houseListEl.innerHTML = "";

  if (computed.length === 0) {
    houseListEl.innerHTML = `
      <div class="empty-state">
        <h3>先录第一套房</h3>
        <p>录入后这里会自动出现排序卡片，方便横向比较和决策。</p>
      </div>
    `;
    return;
  }

  computed.forEach((house) => {
    const card = document.createElement("article");
    card.className = "house-card";
    card.innerHTML = `
      <div class="house-card__head">
        <div>
          <h3>${house.communityName || "未命名房源"}</h3>
          <div class="house-meta">${house.listingName || "待补充楼栋室"} · ${house.viewDate || "未设置日期"}</div>
        </div>
        <span class="pill ${riskToneClass(house.riskLevel)}">${house.decisionLabel}</span>
      </div>
      <div class="house-card__body">
        <div class="score-grid">
          <div class="stat-chip"><span>我的分数</span><strong>${round1(house.myScore)}</strong></div>
          <div class="stat-chip"><span>她的分数</span><strong>${round1(house.partnerScore)}</strong></div>
          <div class="stat-chip"><span>共识分</span><strong>${round1(house.consensusScore)}</strong></div>
          <div class="stat-chip"><span>风险等级</span><strong>${house.riskLevel}</strong></div>
        </div>
        <div>
          <span class="group-label">最终分</span>
          <strong style="display:block;font-size:2rem;margin-top:6px;">${round1(house.finalScore)}</strong>
          <div class="house-meta">
            bonus +${house.bonusScore} / risk -${house.riskPenalty} / 一致性 ${round1(house.consistencyScore)}
          </div>
        </div>
        <div>
          <span class="group-label">风险项</span>
          <div class="risk-list">${renderRiskItems(house)}</div>
        </div>
        <div>
          <span class="group-label">加分项</span>
          <div class="bonus-list">${renderBonusItems(house)}</div>
        </div>
        <p class="house-meta">${house.notes || "暂无补充备注"}</p>
        <div class="card-actions">
          <button class="chip-button" type="button" data-edit="${house.id}">编辑</button>
          <button class="chip-button" type="button" data-duplicate="${house.id}">复制</button>
        </div>
      </div>
    `;
    houseListEl.appendChild(card);
  });

  houseListEl.querySelectorAll("[data-edit]").forEach((button) => {
    button.addEventListener("click", () => {
      editingHouseId = button.dataset.edit;
      renderHouseForm();
      houseFormEl.scrollIntoView({ behavior: "smooth", block: "start" });
    });
  });

  houseListEl.querySelectorAll("[data-duplicate]").forEach((button) => {
    button.addEventListener("click", () => {
      const source = state.houses.find((item) => item.id === button.dataset.duplicate);
      if (!source) return;
      const copy = { ...source, id: crypto.randomUUID(), listingName: `${source.listingName || "房源"} 复制` };
      state.houses.unshift(copy);
      editingHouseId = copy.id;
      persist();
      render();
    });
  });
}

function getEditingHouse() {
  return state.houses.find((item) => item.id === editingHouseId) ?? ensureOneHouse();
}

function ensureOneHouse() {
  if (state.houses.length === 0) {
    const house = DEFAULT_HOUSE();
    state.houses.push(house);
    editingHouseId = house.id;
    persist();
    return house;
  }
  return state.houses[0];
}

function computeAllHouseScores(houses, weights) {
  const validHouses = houses.filter(isHouseStarted);
  const ranges = {};

  METRIC_DEFS.forEach((metric) => {
    if (metric.type === "mapped") return;
    const values = validHouses
      .map((house) => getMetricValue(house, metric.key))
      .filter((value) => Number.isFinite(value));

    ranges[metric.key] = {
      min: values.length ? Math.min(...values) : 0,
      max: values.length ? Math.max(...values) : 100
    };
  });

  return validHouses
    .map((house) => {
      const normalized = {};
      METRIC_DEFS.forEach((metric) => {
        normalized[metric.key] = normalizeMetric(house, metric, ranges);
      });

      const myScore = weightedScore(normalized, weights.me);
      const partnerScore = weightedScore(normalized, weights.partner);
      const consistencyScore = clamp(100 - Math.abs(myScore - partnerScore), 0, 100);
      const consensusScore = 0.45 * myScore + 0.45 * partnerScore + 0.1 * consistencyScore;
      const bonusScore = sumSelected(house.bonusSelections, BONUS_OPTIONS, "score");
      const riskPenalty = sumSelected(house.riskSelections, RISK_OPTIONS, "penalty");
      const finalScore = clamp(consensusScore + bonusScore - riskPenalty, 0, 100);

      return {
        ...house,
        normalized,
        myScore,
        partnerScore,
        consistencyScore,
        consensusScore,
        bonusScore,
        riskPenalty,
        finalScore,
        riskLevel: classifyRisk(riskPenalty),
        decisionLabel: classifyDecision(finalScore)
      };
    })
    .sort((a, b) => b.finalScore - a.finalScore);
}

function isHouseStarted(house) {
  if (!house) return false;
  return Boolean(
    house.communityName ||
      house.listingName ||
      Number(house.totalPrice) > 0 ||
      Number(house.unitPrice) > 0 ||
      Number(house.area) > 0 ||
      house.notes ||
      house.bonusSelections?.length ||
      house.riskSelections?.length
  );
}

function getMetricValue(house, key) {
  if (key === "houseTypeScore") return HOUSE_TYPE_SCORES[house.houseType] ?? 60;
  if (key === "renovationScore") return RENOVATION_SCORES[house.renovation] ?? 60;
  if (key === "noiseScore") return Number(house.noiseScore);
  return Number(house[key]);
}

function normalizeMetric(house, metric, ranges) {
  if (metric.key === "houseTypeScore") return HOUSE_TYPE_SCORES[house.houseType] ?? 60;
  if (metric.key === "renovationScore") return RENOVATION_SCORES[house.renovation] ?? 60;

  const value = getMetricValue(house, metric.key);
  const range = ranges[metric.key];
  if (!Number.isFinite(value)) return 50;
  if (!range || range.max === range.min) {
    return fallbackScore(metric, value);
  }

  if (metric.type === "lower") {
    return clamp((100 * (range.max - value)) / (range.max - range.min), 0, 100);
  }

  return clamp((100 * (value - range.min)) / (range.max - range.min), 0, 100);
}

function fallbackScore(metric, value) {
  if (metric.type === "lower") {
    return clamp(100 - value, 0, 100);
  }

  if (metric.type === "higher") {
    return clamp((value / 10) * 100, 0, 100);
  }

  return clamp(value, 0, 100);
}

function weightedScore(normalized, weights) {
  const totalWeight = Object.values(weights).reduce((sum, value) => sum + Number(value || 0), 0) || 1;
  return METRIC_DEFS.reduce((sum, metric) => {
    const ratio = Number(weights[metric.key] || 0) / totalWeight;
    return sum + normalized[metric.key] * ratio;
  }, 0);
}

function sumSelected(selected, options, scoreKey) {
  return selected.reduce((sum, key) => {
    const option = options.find((item) => item.key === key);
    return sum + (option?.[scoreKey] ?? 0);
  }, 0);
}

function classifyRisk(penalty) {
  if (penalty === 0) return "无风险";
  if (penalty >= 20) return "高风险";
  if (penalty >= 10) return "中风险";
  if (penalty > 0) return "低风险";
  return "无风险";
}

function classifyDecision(score) {
  if (score >= 85) return "强推荐";
  if (score >= 75) return "可重点复看";
  if (score >= 65) return "可谈可比";
  return "谨慎/放弃";
}

function riskToneClass(level) {
  if (level === "高风险") return "bad";
  if (level === "中风险") return "warn";
  return "good";
}

function renderRiskItems(house) {
  if (house.riskSelections.length === 0) {
    return `<span class="risk-item">暂无明显风险</span>`;
  }
  return house.riskSelections
    .map((key) => {
      const risk = RISK_OPTIONS.find((item) => item.key === key);
      return `<span class="risk-item bad">${risk?.label ?? key} -${risk?.penalty ?? 0}</span>`;
    })
    .join("");
}

function renderBonusItems(house) {
  if (house.bonusSelections.length === 0) {
    return `<span class="bonus-item">暂无额外加分</span>`;
  }
  return house.bonusSelections
    .map((key) => {
      const bonus = BONUS_OPTIONS.find((item) => item.key === key);
      return `<span class="bonus-item good">${bonus?.label ?? key} +${bonus?.score ?? 0}</span>`;
    })
    .join("");
}

function round1(value) {
  return Number(value).toFixed(1);
}

function avg(values) {
  return values.reduce((sum, value) => sum + value, 0) / values.length;
}

function clamp(value, min, max) {
  return Math.min(Math.max(value, min), max);
}

ensureOneHouse();
render();
