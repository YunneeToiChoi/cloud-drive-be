package handlers

import (
	"encoding/json"
	"github.com/cloud-drive/api-gateway/internal/clients"
	"github.com/cloud-drive/api-gateway/internal/config"
	"github.com/cloud-drive/api-gateway/internal/middleware"
	"github.com/cloud-drive/proto-definitions/user"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

// AuthHandler xử lý các yêu cầu liên quan đến xác thực
type AuthHandler struct {
	userClient *clients.UserClient
	cfg        *config.Config
}

// LoginRequest là cấu trúc dữ liệu cho yêu cầu đăng nhập
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterRequest là cấu trúc dữ liệu cho yêu cầu đăng ký
type RegisterRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// AuthResponse là cấu trúc dữ liệu cho phản hồi xác thực
type AuthResponse struct {
	Token     string `json:"token"`
	ExpiresIn int64  `json:"expires_in"`
	UserID    string `json:"user_id"`
	Role      string `json:"role"`
}

// NewAuthHandler tạo một handler mới cho xác thực
func NewAuthHandler(userClient *clients.UserClient, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		userClient: userClient,
		cfg:        cfg,
	}
}

// Login xử lý đăng nhập người dùng
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Xác thực người dùng thông qua User Service
	ctx := r.Context()
	userResp, err := h.userClient.Authenticate(ctx, req.Email, req.Password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Tạo token JWT
	token, err := middleware.GenerateToken(userResp.User.Id, userResp.User.Role, h.cfg.JWTSecret, h.cfg.JWTExpiration)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Tạo response
	resp := AuthResponse{
		Token:     token,
		ExpiresIn: int64(h.cfg.JWTExpiration / time.Second),
		UserID:    userResp.User.Id,
		Role:      userResp.User.Role,
	}

	// Gửi response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Register xử lý đăng ký người dùng mới
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Tạo người dùng mới thông qua User Service
	ctx := r.Context()
	userReq := &user.CreateUserRequest{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      "user", // Người dùng mới mặc định có vai trò "user"
	}

	userResp, err := h.userClient.CreateUser(ctx, userReq)
	if err != nil {
		http.Error(w, "Failed to create user: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Tạo token JWT
	token, err := middleware.GenerateToken(userResp.User.Id, "user", h.cfg.JWTSecret, h.cfg.JWTExpiration)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Tạo response
	resp := AuthResponse{
		Token:     token,
		ExpiresIn: int64(h.cfg.JWTExpiration / time.Second),
		UserID:    userResp.User.Id,
		Role:      "user",
	}

	// Gửi response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// RegisterAuthRoutes đăng ký các route cho xác thực
func RegisterAuthRoutes(router *mux.Router, userClient *clients.UserClient, cfg *config.Config) {
	handler := NewAuthHandler(userClient, cfg)

	// Các route cho xác thực
	authRouter := router.PathPrefix("/api/auth").Subrouter()
	authRouter.HandleFunc("/login", handler.Login).Methods("POST")
	authRouter.HandleFunc("/register", handler.Register).Methods("POST")

	// Endpoint kiểm tra token - yêu cầu xác thực
	checkTokenRouter := authRouter.PathPrefix("/check-token").Subrouter()
	checkTokenRouter.Use(middleware.AuthMiddleware(cfg))
	checkTokenRouter.HandleFunc("", handler.CheckToken).Methods("GET")
}

// CheckToken xác nhận token hiện tại có hợp lệ không và trả về thông tin người dùng từ token
func (h *AuthHandler) CheckToken(w http.ResponseWriter, r *http.Request) {
	// Lấy thông tin claims từ context (đã được thêm bởi middleware)
	claims, ok := r.Context().Value("claims").(*middleware.Claims)
	if !ok {
		http.Error(w, "Không tìm thấy thông tin xác thực", http.StatusUnauthorized)
		return
	}

	// Tạo response chứa thông tin về token
	type TokenResponse struct {
		Valid     bool      `json:"valid"`
		UserID    string    `json:"user_id"`
		Role      string    `json:"role"`
		ExpiresAt time.Time `json:"expires_at"`
	}

	response := TokenResponse{
		Valid:     true,
		UserID:    claims.UserID,
		Role:      claims.Role,
		ExpiresAt: time.Unix(claims.ExpiresAt, 0),
	}

	// Gửi response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
