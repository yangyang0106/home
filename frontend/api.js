// 生产环境通过 Nginx 反向代理 `/api` 到 Go 后端，
// 前端统一走相对路径，避免写死本机地址后在服务器上访问失效。
const API_BASE = "/api/v1";
const AUTH_TOKEN_KEY = "home-decision-auth-token";

window.homeApi = {
  householdId: "",
  token: localStorage.getItem(AUTH_TOKEN_KEY) || "",
  setToken(token) {
    this.token = token || "";
    if (this.token) localStorage.setItem(AUTH_TOKEN_KEY, this.token);
    else localStorage.removeItem(AUTH_TOKEN_KEY);
  },
  async register(payload) {
    return request("/auth/register", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  },
  async login(payload) {
    return request("/auth/login", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  },
  async logout() {
    return request("/auth/logout", { method: "POST" });
  },
  async getMe() {
    return request("/auth/me");
  },
  async linkPartner(partnerLinkCode) {
    return request("/auth/link", {
      method: "POST",
      body: JSON.stringify({ partnerLinkCode })
    });
  },
  async getAdminUsers() {
    return request("/admin/users");
  },
  async setUserAdmin(userID, isAdmin) {
    return request(`/admin/users/${userID}/admin`, {
      method: "PUT",
      body: JSON.stringify({ isAdmin })
    });
  },
  async getDashboard() {
    return request(`/households/${this.householdId}/dashboard`);
  },
  async saveWeights(profiles) {
    return request(`/households/${this.householdId}/weights`, {
      method: "PUT",
      body: JSON.stringify({ profiles })
    });
  },
  async createHouse(house) {
    return request(`/households/${this.householdId}/houses`, {
      method: "POST",
      body: JSON.stringify(house)
    });
  },
  async updateHouse(house) {
    return request(`/households/${this.householdId}/houses/${house.id}`, {
      method: "PUT",
      body: JSON.stringify(house)
    });
  },
  async deleteHouse(id) {
    return request(`/households/${this.householdId}/houses/${id}`, {
      method: "DELETE"
    });
  }
};

async function request(path, options = {}) {
  const headers = {
    "Content-Type": "application/json",
    ...(window.homeApi.token ? { Authorization: `Bearer ${window.homeApi.token}` } : {}),
    ...(options.headers || {})
  };
  const response = await fetch(`${API_BASE}${path}`, {
    headers,
    ...options
  });
  if (!response.ok) {
    const data = await response.json().catch(() => ({}));
    throw new Error(data.error || "request failed");
  }
  return response.json();
}
