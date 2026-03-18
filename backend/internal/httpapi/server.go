package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"home-decision/backend/internal/model"
	"home-decision/backend/internal/service"
	"home-decision/backend/internal/store"
)

type Server struct {
	store   store.Store
	scoring *service.ScoringService
	auth    *service.AuthService
	origin  string
}

func NewServer(s store.Store, scoring *service.ScoringService, auth *service.AuthService, allowedOrigin string) *Server {
	return &Server{store: s, scoring: scoring, auth: auth, origin: allowedOrigin}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/api/v1/meta", s.handleMeta)
	mux.HandleFunc("/api/v1/auth/register", s.handleRegister)
	mux.HandleFunc("/api/v1/auth/login", s.handleLogin)
	mux.HandleFunc("/api/v1/auth/logout", s.handleLogout)
	mux.HandleFunc("/api/v1/auth/me", s.handleMe)
	mux.HandleFunc("/api/v1/auth/link", s.handleLink)
	mux.HandleFunc("/api/v1/admin/users", s.handleAdminUsers)
	mux.HandleFunc("/api/v1/admin/users/", s.handleAdminUserAction)
	mux.HandleFunc("/api/v1/households/", s.handleHouseholds)
	return s.withCORS(s.withJSON(mux))
}

func (s *Server) withJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		next.ServeHTTP(w, r)
	})
}

func (s *Server) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", s.origin)
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func (s *Server) handleMeta(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.store.GetMeta())
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req struct {
		LoginID     string `json:"loginId"`
		Password    string `json:"password"`
		DisplayName string `json:"displayName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	user, token, err := s.auth.Register(req.LoginID, req.Password, req.DisplayName)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"user": user, "token": token})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req struct {
		LoginID  string `json:"loginId"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	user, token, err := s.auth.Login(req.LoginID, req.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": user, "token": token})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	_ = s.auth.Logout(bearerToken(r))
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	profile, err := s.auth.ProfileByToken(bearerToken(r))
	if err != nil || profile == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	writeJSON(w, http.StatusOK, profile)
}

func (s *Server) handleLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req struct {
		PartnerLinkCode string `json:"partnerLinkCode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	profile, err := s.auth.LinkPartner(bearerToken(r), req.PartnerLinkCode)
	if err != nil {
		if err.Error() == "unauthorized" {
			writeError(w, http.StatusUnauthorized, err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, profile)
}

func (s *Server) handleAdminUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	items, err := s.auth.AdminUsers(bearerToken(r))
	if err != nil {
		if err.Error() == "forbidden" {
			writeError(w, http.StatusForbidden, err.Error())
			return
		}
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) handleAdminUserAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/users/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 2 || parts[1] != "admin" {
		writeError(w, http.StatusNotFound, "resource not found")
		return
	}
	var req struct {
		IsAdmin bool `json:"isAdmin"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := s.auth.SetAdmin(bearerToken(r), parts[0], req.IsAdmin); err != nil {
		switch err.Error() {
		case "unauthorized":
			writeError(w, http.StatusUnauthorized, err.Error())
		case "forbidden":
			writeError(w, http.StatusForbidden, err.Error())
		default:
			writeError(w, http.StatusBadRequest, err.Error())
		}
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleHouseholds(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/households/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		writeError(w, http.StatusNotFound, "resource not found")
		return
	}
	householdID := parts[0]
	user, err := s.authProfileForHousehold(bearerToken(r), householdID)
	if err != nil {
		if err.Error() == "forbidden" {
			writeError(w, http.StatusForbidden, err.Error())
			return
		}
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	resource := parts[1]

	switch resource {
	case "dashboard":
		s.handleDashboard(w, r, householdID)
	case "weights":
		s.handleWeights(w, r, householdID)
	case "houses":
		s.handleHouses(w, r, householdID, parts[2:])
	default:
		writeError(w, http.StatusNotFound, "resource not found")
	}
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request, householdID string) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	profiles, err := s.store.GetWeights(householdID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	houses, err := s.store.ListHouses(householdID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	dashboard := service.AssembleDashboard(householdID, profiles, houses, s.store.GetMeta())
	writeJSON(w, http.StatusOK, dashboard)
}

func (s *Server) handleWeights(w http.ResponseWriter, r *http.Request, householdID string) {
	switch r.Method {
	case http.MethodGet:
		profiles, err := s.store.GetWeights(householdID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": profiles})
	case http.MethodPut:
		var req struct {
			Profiles []model.WeightProfile `json:"profiles"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		if err := s.store.SaveWeights(householdID, req.Profiles); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handleHouses(w http.ResponseWriter, r *http.Request, householdID string, tail []string) {
	if len(tail) == 0 {
		switch r.Method {
		case http.MethodGet:
			houses, err := s.store.ListHouses(householdID)
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"items": houses})
		case http.MethodPost:
			var house model.House
			if err := json.NewDecoder(r.Body).Decode(&house); err != nil {
				writeError(w, http.StatusBadRequest, "invalid json")
				return
			}
			created, err := s.store.CreateHouse(householdID, house)
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			writeJSON(w, http.StatusCreated, created)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	houseID := tail[0]
	switch r.Method {
	case http.MethodGet:
		house, err := s.store.GetHouse(householdID, houseID)
		if err != nil {
			writeNotFound(w, err)
			return
		}
		writeJSON(w, http.StatusOK, house)
	case http.MethodPut:
		var house model.House
		if err := json.NewDecoder(r.Body).Decode(&house); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		updated, err := s.store.UpdateHouse(householdID, houseID, house)
		if err != nil {
			writeNotFound(w, err)
			return
		}
		writeJSON(w, http.StatusOK, updated)
	case http.MethodDelete:
		if err := s.store.DeleteHouse(householdID, houseID); err != nil {
			writeNotFound(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func writeNotFound(w http.ResponseWriter, err error) {
	if errors.Is(err, store.ErrHouseNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeError(w, http.StatusNotFound, err.Error())
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{"error": message})
}

func bearerToken(r *http.Request) string {
	value := r.Header.Get("Authorization")
	if value == "" || !strings.HasPrefix(value, "Bearer ") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(value, "Bearer "))
}

func (s *Server) authProfileForHousehold(token, householdID string) (*model.User, error) {
	user, err := s.store.FindUserBySessionToken(token)
	if err != nil || user == nil {
		return nil, err
	}
	if user.IsAdmin {
		return user, nil
	}
	myHouseholdID, err := s.store.FindHouseholdIDByUserID(user.ID)
	if err != nil {
		return nil, err
	}
	if myHouseholdID != householdID {
		return nil, errors.New("forbidden")
	}
	return user, nil
}
