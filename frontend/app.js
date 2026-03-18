const api = window.homeApi;

const VIEW_IDS = ["dashboard", "list", "create", "edit", "compare", "weights", "account", "admin"];

let state = {
  meta: null,
  weights: [],
  houses: [],
  summary: null
};
let authProfile = null;
let activeWeightProfile = "me";
let editHouseId = null;
let createDraft = createEmptyHouse();
let compareSelection = new Set();
let listSearchKeyword = "";
let listSortMode = "final-desc";
let adminUsers = [];

const authPanelEl = document.getElementById("auth-panel");
const appMainEl = document.getElementById("app-main");
const topNavEl = document.getElementById("top-nav");
const summaryCardsEl = document.getElementById("summary-cards");
const topPicksEl = document.getElementById("top-picks");
const weightTabsEl = document.getElementById("weight-tabs");
const weightEditorEl = document.getElementById("weight-editor");
const houseListEl = document.getElementById("house-list");
const createFormEl = document.getElementById("create-form");
const editFormEl = document.getElementById("edit-form");
const editTitleEl = document.getElementById("edit-title");
const statusTextEl = document.getElementById("status-text");
const comparePickerEl = document.getElementById("compare-picker");
const compareResultsEl = document.getElementById("compare-results");
const accountProfileEl = document.getElementById("account-profile");
const adminUsersEl = document.getElementById("admin-users");
const weightRowTemplate = document.getElementById("weight-row-template");
const listSearchEl = document.getElementById("list-search");
const listSortEl = document.getElementById("list-sort");

document.getElementById("reload-dashboard").addEventListener("click", loadDashboard);
document.getElementById("reset-create").addEventListener("click", () => {
  createDraft = createEmptyHouse();
  renderCreateForm();
});
document.getElementById("back-dashboard").addEventListener("click", () => switchView("dashboard"));
document.getElementById("logout-button").addEventListener("click", logout);

document.getElementById("login-form").addEventListener("submit", async (event) => {
  event.preventDefault();
  const loginId = document.getElementById("login-id").value.trim();
  const password = document.getElementById("login-password").value;
  const result = await api.login({ loginId, password });
  api.setToken(result.token);
  await bootstrapAuthedApp();
});

document.getElementById("register-form").addEventListener("submit", async (event) => {
  event.preventDefault();
  const displayName = document.getElementById("register-name").value.trim();
  const loginId = document.getElementById("register-id").value.trim();
  const password = document.getElementById("register-password").value;
  const result = await api.register({ displayName, loginId, password });
  api.setToken(result.token);
  await bootstrapAuthedApp();
});

document.getElementById("link-form").addEventListener("submit", async (event) => {
  event.preventDefault();
  const partnerLinkCode = document.getElementById("partner-link-code").value.trim();
  if (!partnerLinkCode) return;
  authProfile = await api.linkPartner(partnerLinkCode);
  api.householdId = authProfile.householdId;
  document.getElementById("partner-link-code").value = "";
  await loadDashboard();
  renderAccount();
});

listSearchEl.addEventListener("input", (event) => {
  listSearchKeyword = event.target.value.trim().toLowerCase();
  renderHouseList();
});

listSortEl.addEventListener("change", (event) => {
  listSortMode = event.target.value;
  renderHouseList();
});

document.querySelectorAll("[data-view]").forEach((button) => {
  button.addEventListener("click", () => switchView(button.dataset.view));
});

function createEmptyHouse() {
  return {
    id: crypto.randomUUID(),
    householdId: api.householdId || "",
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
  };
}

async function bootstrapAuthedApp() {
  authProfile = await api.getMe();
  api.householdId = authProfile.householdId;
  createDraft = createEmptyHouse();
  authPanelEl.classList.add("hidden");
  appMainEl.classList.remove("hidden");
  topNavEl.classList.remove("hidden");
  syncAdminNav();
  switchView("dashboard");
  renderAccount();
  if (authProfile.user.isAdmin) {
    await loadAdminUsers();
  }
  await loadDashboard();
}

async function logout() {
  try {
    await api.logout();
  } catch {}
  api.setToken("");
  authProfile = null;
  state = { meta: null, weights: [], houses: [], summary: null };
  compareSelection = new Set();
  authPanelEl.classList.remove("hidden");
  appMainEl.classList.add("hidden");
  topNavEl.classList.add("hidden");
}

async function loadDashboard() {
  if (!api.householdId) return;
  statusTextEl.textContent = "正在同步后端数据...";
  const dashboard = await api.getDashboard();
  state = dashboard;
  syncCompareSelection();
  if (!editHouseId && dashboard.houses.length) {
    editHouseId = dashboard.houses[0].id;
  }
  render();
  statusTextEl.textContent = `已连接家庭 · ${dashboard.householdId}`;
}

function syncCompareSelection() {
  const existingIds = new Set(state.houses.map((house) => house.id));
  compareSelection = new Set(Array.from(compareSelection).filter((id) => existingIds.has(id)));
}

function switchView(view) {
  if (view === "admin" && !authProfile?.user?.isAdmin) {
    view = "dashboard";
  }
  document.querySelectorAll(".nav-button").forEach((button) => {
    button.classList.toggle("active", button.dataset.view === view);
  });
  VIEW_IDS.forEach((id) => {
    const node = document.getElementById(`view-${id}`);
    if (node) node.classList.toggle("active", id === view);
  });
}

function render() {
  renderSummary();
  renderHouseList();
  renderCreateForm();
  renderEditForm();
  renderCompare();
  renderWeightTabs();
  renderWeightEditor();
  renderAccount();
  renderAdmin();
}

function renderSummary() {
  summaryCardsEl.innerHTML = "";
  topPicksEl.innerHTML = "";
  const summary = state.summary || { count: 0, bestFinalScore: 0, averageConsensus: 0 };
  [
    { label: "总房源数", value: summary.count, suffix: "套" },
    { label: "当前最高分", value: round1(summary.bestFinalScore || 0), suffix: "分" },
    { label: "平均共识分", value: round1(summary.averageConsensus || 0), suffix: "分" }
  ].forEach((card) => {
    const el = document.createElement("article");
    el.className = "summary-card";
    el.innerHTML = `<span>${card.label}</span><strong>${card.value}</strong><span>${card.suffix}</span>`;
    summaryCardsEl.appendChild(el);
  });
  if (summary.topChoice) topPicksEl.appendChild(buildTopPick("当前第一", summary.topChoice, "good"));
  if (summary.runnerUp) topPicksEl.appendChild(buildTopPick("备选第二", summary.runnerUp, "warn"));
}

function buildTopPick(title, item, tone) {
  const node = document.createElement("article");
  node.className = "top-pick";
  node.innerHTML = `
    <div class="top-pick__head">
      <div><span>${title}</span><h3>${item.communityName || "未命名房源"}</h3></div>
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
  houseListEl.innerHTML = "";
  const visibleHouses = getVisibleHouses();
  if (!state.houses.length) {
    houseListEl.innerHTML = `<div class="empty-state"><h3>暂无房源</h3><p>去“新建”页录入第一套房，然后再回来统一管理和对比。</p></div>`;
    return;
  }
  if (!visibleHouses.length) {
    houseListEl.innerHTML = `<div class="empty-state"><h3>没有匹配结果</h3><p>换个关键词试试，或者切换排序方式看看。</p></div>`;
    return;
  }

  visibleHouses.forEach((house) => {
    const checked = compareSelection.has(house.id);
    const card = document.createElement("article");
    card.className = "house-card";
    const communityName = highlightMatch(house.communityName || "未命名房源");
    const listingName = highlightMatch(house.listingName || "待补充楼栋室");
    const notes = highlightMatch(house.notes || "");
    const metaParts = [listingName, house.viewDate || ""].filter(Boolean);
    card.innerHTML = `
      <div class="house-card__head">
        <div>
          <h3>${communityName}</h3>
          <div class="house-meta">${metaParts.join(" · ")}</div>
        </div>
        <span class="pill ${riskToneClass(house.riskLevel)}">${house.decisionLabel}</span>
      </div>
      <div class="house-card__body">
        <div class="score-grid">
          <div class="stat-chip"><span>我的分数</span><strong>${round1(house.myScore)}</strong></div>
          <div class="stat-chip"><span>她的分数</span><strong>${round1(house.partnerScore)}</strong></div>
          <div class="stat-chip"><span>共识分</span><strong>${round1(house.consensusScore)}</strong></div>
          <div class="stat-chip"><span>最终分</span><strong>${round1(house.finalScore)}</strong></div>
        </div>
        <div class="house-meta">风险等级：${house.riskLevel} · bonus +${house.bonusScore} / risk -${house.riskPenalty}</div>
        ${notes ? `<div class="house-notes">${notes}</div>` : ""}
        <div class="card-actions">
          <button class="chip-button" type="button" data-edit="${house.id}">编辑</button>
          <button class="chip-button ${checked ? "active-chip" : ""}" type="button" data-compare="${house.id}">${checked ? "已加入对比" : "加入对比"}</button>
        </div>
      </div>
    `;
    houseListEl.appendChild(card);
  });

  houseListEl.querySelectorAll("[data-edit]").forEach((button) => {
    button.addEventListener("click", () => {
      editHouseId = button.dataset.edit;
      renderEditForm();
      switchView("edit");
    });
  });
  houseListEl.querySelectorAll("[data-compare]").forEach((button) => {
    button.addEventListener("click", () => {
      const id = button.dataset.compare;
      if (compareSelection.has(id)) compareSelection.delete(id);
      else compareSelection.add(id);
      renderHouseList();
      renderCompare();
    });
  });
}

function getVisibleHouses() {
  const filtered = state.houses.filter((house) => {
    if (!listSearchKeyword) return true;
    const text = [house.communityName, house.listingName, house.notes, house.orientation, house.floor]
      .filter(Boolean)
      .join(" ")
      .toLowerCase();
    return text.includes(listSearchKeyword);
  });
  return filtered.sort((a, b) => compareHouse(a, b, listSortMode));
}

function compareHouse(a, b, mode) {
  switch (mode) {
    case "final-asc": return Number(a.finalScore || 0) - Number(b.finalScore || 0);
    case "date-desc": return parseDate(b.viewDate) - parseDate(a.viewDate);
    case "date-asc": return parseDate(a.viewDate) - parseDate(b.viewDate);
    case "price-asc": return Number(a.totalPrice || 0) - Number(b.totalPrice || 0);
    case "price-desc": return Number(b.totalPrice || 0) - Number(a.totalPrice || 0);
    default: return Number(b.finalScore || 0) - Number(a.finalScore || 0);
  }
}

function parseDate(value) {
  if (!value) return 0;
  const time = new Date(value).getTime();
  return Number.isNaN(time) ? 0 : time;
}

function highlightMatch(text) {
  const value = String(text || "");
  if (!listSearchKeyword) return escapeHtml(value);
  const regex = new RegExp(`(${escapeRegExp(listSearchKeyword)})`, "ig");
  return escapeHtml(value).replace(regex, '<mark class="search-hit">$1</mark>');
}

function escapeHtml(value) {
  return value
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}

function escapeRegExp(value) {
  return String(value).replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

function renderCompare() {
  renderComparePicker();
  renderCompareResults();
}

function renderComparePicker() {
  comparePickerEl.innerHTML = "";
  if (!state.houses.length) {
    comparePickerEl.innerHTML = `<div class="empty-state"><h3>暂无可对比房源</h3><p>先去新建页录入几套房，再回来做横向比较。</p></div>`;
    return;
  }
  state.houses.forEach((house) => {
    const item = document.createElement("label");
    item.className = "choice-item";
    item.innerHTML = `
      <div>
        <div>${house.communityName || "未命名房源"}</div>
        <small>${house.listingName || "待补充楼栋室"} · 最终分 ${round1(house.finalScore)}</small>
      </div>
      <input type="checkbox" ${compareSelection.has(house.id) ? "checked" : ""} />
    `;
    item.querySelector("input").addEventListener("change", (event) => {
      if (event.target.checked) compareSelection.add(house.id);
      else compareSelection.delete(house.id);
      renderHouseList();
      renderCompareResults();
    });
    comparePickerEl.appendChild(item);
  });
}

function renderCompareResults() {
  compareResultsEl.innerHTML = "";
  const picked = state.houses.filter((house) => compareSelection.has(house.id));
  if (picked.length < 2) {
    compareResultsEl.innerHTML = `<div class="empty-state"><h3>至少选 2 套</h3><p>勾选你想比的房源，系统会按核心结果并排展示。</p></div>`;
    return;
  }
  const table = document.createElement("div");
  table.className = "compare-table";
  const rows = [
    { label: "最终分", key: "finalScore", format: round1 },
    { label: "我的分", key: "myScore", format: round1 },
    { label: "她的分", key: "partnerScore", format: round1 },
    { label: "共识分", key: "consensusScore", format: round1 },
    { label: "风险等级", key: "riskLevel", format: (v) => v },
    { label: "总价", key: "totalPrice", format: (v) => `${v} 万` },
    { label: "通勤", key: "commuteTime", format: (v) => `${v} 分钟` }
  ];
  const header = document.createElement("div");
  header.className = "compare-row compare-row--head";
  header.innerHTML = `<div class="compare-cell compare-cell--label">指标</div>${picked.map((house) => `<div class="compare-cell"><strong>${house.communityName || "未命名房源"}</strong><small>${house.listingName || ""}</small></div>`).join("")}`;
  table.appendChild(header);
  rows.forEach((row) => {
    const line = document.createElement("div");
    line.className = "compare-row";
    line.innerHTML = `<div class="compare-cell compare-cell--label">${row.label}</div>${picked.map((house) => `<div class="compare-cell">${row.format(house[row.key])}</div>`).join("")}`;
    table.appendChild(line);
  });
  compareResultsEl.appendChild(table);
}

function renderWeightTabs() {
  weightTabsEl.innerHTML = "";
  state.weights.forEach((profile) => {
    const button = document.createElement("button");
    button.type = "button";
    button.className = `weight-tab${activeWeightProfile === profile.role ? " active" : ""}`;
    button.innerHTML = `<strong>${profile.label}</strong><div class="hint-text">${profile.role}</div>`;
    button.addEventListener("click", () => {
      activeWeightProfile = profile.role;
      renderWeightTabs();
      renderWeightEditor();
    });
    weightTabsEl.appendChild(button);
  });
}

function renderWeightEditor() {
  weightEditorEl.innerHTML = "";
  const profile = state.weights.find((item) => item.role === activeWeightProfile);
  if (!profile || !state.meta) return;
  state.meta.metrics.forEach((metric) => {
    const row = weightRowTemplate.content.firstElementChild.cloneNode(true);
    const input = row.querySelector("input");
    const output = row.querySelector("output");
    row.querySelector(".metric-name").textContent = metric.label;
    row.querySelector(".metric-note").textContent = metric.note;
    input.value = profile.weights[metric.key] ?? 0;
    output.value = `${input.value}%`;
    input.addEventListener("change", async (event) => {
      profile.weights[metric.key] = Number(event.target.value);
      output.value = `${event.target.value}%`;
      await api.saveWeights(state.weights);
      await loadDashboard();
    });
    weightEditorEl.appendChild(row);
  });
}

function renderCreateForm() {
  createDraft.householdId = api.householdId;
  renderHouseForm(createFormEl, createDraft, {
    submitLabel: "创建房源",
    onSubmit: async () => {
      await api.createHouse(createDraft);
      createDraft = createEmptyHouse();
      await loadDashboard();
      switchView("list");
    },
    secondaryLabel: "清空",
    onSecondary: () => {
      createDraft = createEmptyHouse();
      renderCreateForm();
    }
  });
}

function renderEditForm() {
  const house = getEditHouse();
  editFormEl.innerHTML = "";
  if (!house) {
    editTitleEl.textContent = "暂无可编辑房源";
    editFormEl.innerHTML = `<div class="empty-state"><h3>还没有房源</h3><p>先去新建页录入一套房，再来这里编辑。</p></div>`;
    return;
  }
  editTitleEl.textContent = `${house.communityName || "编辑房源"} · ${house.listingName || ""}`;
  renderHouseForm(editFormEl, house, {
    submitLabel: "保存修改",
    onSubmit: async () => {
      await api.updateHouse(house);
      await loadDashboard();
      switchView("list");
    },
    secondaryLabel: "删除房源",
    onSecondary: async () => {
      await api.deleteHouse(house.id);
      editHouseId = null;
      await loadDashboard();
      switchView("list");
    }
  });
}

function renderHouseForm(container, house, options) {
  container.innerHTML = "";
  const groups = [
    ["基础信息", [["communityName", "小区名"], ["listingName", "房源名/楼栋室"], ["viewDate", "看房日期", "date"], ["totalPrice", "总价", "number"], ["unitPrice", "单价", "number"], ["area", "面积", "number"], ["houseAge", "房龄", "number"], ["floor", "楼层"], ["orientation", "朝向"]]],
    ["通勤与生活", [["commuteTime", "通勤", "number"], ["metroTime", "地铁", "number"], ["monthlyFee", "物业费", "number"], ["efficiencyRate", "得房率", "number"], ["livingConvenience", "生活便利度", "number"]]],
    ["居住体验", [["lightScore", "采光", "number"], ["noiseScore", "噪音", "number"], ["layoutScore", "户型", "number"], ["propertyScore", "物业观感", "number"], ["communityScore", "小区环境", "number"], ["comfortScore", "舒适感", "number"], ["parkingScore", "停车便利", "number"]]]
  ];
  groups.forEach(([title, fields]) => {
    const section = document.createElement("section");
    section.className = "form-group";
    section.innerHTML = `<h3>${title}</h3>`;
    const grid = document.createElement("div");
    grid.className = "fields-grid two-col";
    fields.forEach(([key, label, type = "text"]) => grid.appendChild(buildField(house, key, label, type)));
    section.appendChild(grid);
    container.appendChild(section);
  });
  container.appendChild(buildSelectSection(house));
  container.appendChild(buildChoiceSection("附加加分项", "bonusSelections", state.meta.bonusOptions, house));
  container.appendChild(buildChoiceSection("风险扣分项", "riskSelections", state.meta.riskOptions, house));
  const note = document.createElement("section");
  note.className = "form-group";
  note.innerHTML = `<h3>备注</h3><div class="field"><textarea rows="4">${house.notes || ""}</textarea></div>`;
  note.querySelector("textarea").addEventListener("input", (event) => {
    house.notes = event.target.value;
  });
  container.appendChild(note);
  const actions = document.createElement("div");
  actions.className = "form-actions";
  actions.innerHTML = `<button type="button" class="primary-button">${options.submitLabel}</button><button type="button" class="secondary-button">${options.secondaryLabel}</button>`;
  actions.children[0].addEventListener("click", options.onSubmit);
  actions.children[1].addEventListener("click", options.onSecondary);
  container.appendChild(actions);
}

function buildField(house, key, label, type) {
  const wrap = document.createElement("div");
  wrap.className = "field";
  wrap.innerHTML = `<label>${label}</label><input type="${type}" value="${house[key] ?? ""}" />`;
  wrap.querySelector("input").addEventListener("input", (event) => {
    house[key] = type === "number" ? Number(event.target.value) : event.target.value;
  });
  return wrap;
}

function buildSelectSection(house) {
  const section = document.createElement("section");
  section.className = "form-group";
  section.innerHTML = `<h3>类型设置</h3>`;
  const grid = document.createElement("div");
  grid.className = "fields-grid two-col";
  grid.appendChild(buildSelect(house, "houseType", "房屋类型", Object.keys(state.meta.houseTypeScores)));
  grid.appendChild(buildSelect(house, "renovation", "装修情况", Object.keys(state.meta.renovationScores)));
  section.appendChild(grid);
  return section;
}

function buildSelect(house, key, label, options) {
  const wrap = document.createElement("div");
  wrap.className = "field";
  const select = document.createElement("select");
  options.forEach((option) => {
    const node = document.createElement("option");
    node.value = option;
    node.textContent = option;
    select.appendChild(node);
  });
  select.value = house[key];
  select.addEventListener("change", (event) => {
    house[key] = event.target.value;
  });
  wrap.innerHTML = `<label>${label}</label>`;
  wrap.appendChild(select);
  return wrap;
}

function buildChoiceSection(title, key, options, house) {
  const section = document.createElement("section");
  section.className = "form-group";
  section.innerHTML = `<h3>${title}</h3>`;
  const grid = document.createElement("div");
  grid.className = "choice-grid";
  options.forEach((option) => {
    const item = document.createElement("label");
    item.className = "choice-item";
    const selected = house[key]?.includes(option.key);
    item.innerHTML = `<div><div>${option.label}</div><small>${option.score ? `加分 +${option.score}` : `${option.level}风险 -${option.penalty}`}</small></div><input type="checkbox" ${selected ? "checked" : ""} />`;
    item.querySelector("input").addEventListener("change", (event) => {
      const set = new Set(house[key] || []);
      if (event.target.checked) set.add(option.key);
      else set.delete(option.key);
      house[key] = Array.from(set);
    });
    grid.appendChild(item);
  });
  section.appendChild(grid);
  return section;
}

function getEditHouse() {
  return state.houses.find((house) => house.id === editHouseId) || null;
}

function renderAccount() {
  if (!authProfile) {
    accountProfileEl.innerHTML = "";
    return;
  }
  const members = authProfile.members.length
    ? authProfile.members.map((member) => `<div class="account-member"><strong>${member.displayName}</strong><span>${member.loginId}</span><small>ID: ${member.linkCode}</small></div>`).join("")
    : `<p class="hint-text">当前还没有关联到另一半。</p>`;
  accountProfileEl.innerHTML = `
    <div class="account-card">
      <div><span class="hint-text">显示名</span><strong>${authProfile.user.displayName}</strong></div>
      <div><span class="hint-text">登录账号</span><strong>${authProfile.user.loginId}</strong></div>
      <div><span class="hint-text">我的唯一 ID</span><strong>${authProfile.user.linkCode}</strong></div>
      <div><span class="hint-text">当前家庭 ID</span><strong>${authProfile.householdId}</strong></div>
    </div>
    <div class="account-members">
      <h3>家庭成员</h3>
      ${members}
    </div>
  `;
}

async function loadAdminUsers() {
  if (!authProfile?.user?.isAdmin) {
    adminUsers = [];
    return;
  }
  const result = await api.getAdminUsers();
  adminUsers = result.items || [];
}

function renderAdmin() {
  adminUsersEl.innerHTML = "";
  if (!authProfile?.user?.isAdmin) {
    adminUsersEl.innerHTML = `<div class="empty-state"><h3>无权限</h3><p>当前账号不是管理员，无法查看账号总管理页面。</p></div>`;
    return;
  }
  if (!adminUsers.length) {
    adminUsersEl.innerHTML = `<div class="empty-state"><h3>暂无账号</h3><p>等有人注册后，这里会显示全部账号。</p></div>`;
    return;
  }

  adminUsers.forEach((user) => {
    const item = document.createElement("article");
    item.className = "account-member";
    item.innerHTML = `
      <strong>${user.displayName}</strong>
      <span>${user.loginId}</span>
      <small>唯一 ID: ${user.linkCode}</small>
      <small>家庭: ${user.householdId}</small>
      <div class="card-actions">
        <span class="pill ${user.isAdmin ? "good" : "warn"}">${user.isAdmin ? "管理员" : "普通账号"}</span>
        <button class="chip-button" type="button">${user.isAdmin ? "取消管理员" : "设为管理员"}</button>
      </div>
    `;
    item.querySelector("button").addEventListener("click", async () => {
      await api.setUserAdmin(user.id, !user.isAdmin);
      await loadAdminUsers();
      renderAdmin();
    });
    adminUsersEl.appendChild(item);
  });
}

function syncAdminNav() {
  const adminButton = document.querySelector('[data-view="admin"]');
  if (!adminButton) return;
  adminButton.classList.toggle("hidden", !authProfile?.user?.isAdmin);
}

function riskToneClass(level) {
  if (level === "高风险") return "bad";
  if (level === "中风险") return "warn";
  return "good";
}

function round1(value) {
  return Number(value || 0).toFixed(1);
}

(async function init() {
  if (api.token) {
    try {
      await bootstrapAuthedApp();
      return;
    } catch {
      api.setToken("");
    }
  }
  authPanelEl.classList.remove("hidden");
  appMainEl.classList.add("hidden");
  topNavEl.classList.add("hidden");
})();
