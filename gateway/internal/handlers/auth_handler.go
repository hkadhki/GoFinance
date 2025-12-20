package handlers

import (
	"encoding/json"
	"net/http"

	authv1 "gateway/auth/v1"
)

type AuthHandler struct {
	client authv1.AuthServiceClient
}

func NewAuthHandler(c authv1.AuthServiceClient) *AuthHandler {
	return &AuthHandler{
		client: c,
	}
}

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string `json:"token"`
}

// Register godoc
// @Summary Register
// @Description Register new user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body authRequest true "Register request"
// @Success 200 {object} authResponse
// @Failure 400 {object} map[string]string
// @Router /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	resp, err := h.client.Register(
		r.Context(),
		&authv1.RegisterRequest{
			Email:    req.Email,
			Password: req.Password,
		},
	)
	if err != nil {
		http.Error(w, grpcToHTTP(err), http.StatusUnauthorized)
		return
	}

	writeJSON(w, http.StatusOK, authResponse{Token: resp.AccessToken})
}

// Login godoc
// @Summary Login
// @Description Login user and get JWT
// @Tags auth
// @Accept json
// @Produce json
// @Param request body authRequest true "Login request"
// @Success 200 {object} authResponse
// @Failure 401 {object} map[string]string
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	resp, err := h.client.Login(
		r.Context(),
		&authv1.LoginRequest{
			Email:    req.Email,
			Password: req.Password,
		},
	)
	if err != nil {
		http.Error(w, grpcToHTTP(err), http.StatusUnauthorized)
		return
	}

	writeJSON(w, http.StatusOK, authResponse{Token: resp.AccessToken})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func grpcToHTTP(err error) string {
	return err.Error()
}
