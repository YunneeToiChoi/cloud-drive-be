package middleware

import (
	"context"
	"errors"
	"github.com/cloud-drive/api-gateway/internal/config"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"strings"
	"time"
)

// Claims là cấu trúc dữ liệu cho JWT claims
type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.StandardClaims
}

// AuthMiddleware tạo middleware xác thực JWT
func AuthMiddleware(cfg *config.Config) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Lấy token từ Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header is required", http.StatusUnauthorized)
				return
			}

			// Kiểm tra định dạng "Bearer <token>"
			bearerToken := strings.Split(authHeader, " ")
			if len(bearerToken) != 2 || strings.ToLower(bearerToken[0]) != "bearer" {
				http.Error(w, "Invalid token format", http.StatusUnauthorized)
				return
			}

			// Trích xuất token
			tokenString := bearerToken[1]

			// Xác thực token
			claims, err := validateToken(tokenString, cfg.JWTSecret)
			if err != nil {
				http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}

			// Thêm claims vào context để các handler có thể sử dụng
			ctx := context.WithValue(r.Context(), "claims", claims)

			// Gọi handler tiếp theo với context đã cập nhật
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// validateToken xác thực JWT token và trả về claims
func validateToken(tokenString string, secretKey string) (*Claims, error) {
	// Phân tích token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Đảm bảo đúng phương thức mã hóa
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secretKey), nil
	})

	// Kiểm tra lỗi khi phân tích
	if err != nil {
		return nil, err
	}

	// Xác thực claims
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// GenerateToken tạo JWT token mới
func GenerateToken(userID string, role string, secretKey string, expirationTime time.Duration) (string, error) {
	// Tạo claims
	claims := &Claims{
		UserID: userID,
		Role:   role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(expirationTime).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}

	// Tạo token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Ký token
	return token.SignedString([]byte(secretKey))
}

// GetUserIDFromContext lấy user ID từ JWT claims trong context
func GetUserIDFromContext(ctx context.Context) (string, error) {
	claims, ok := ctx.Value("claims").(*Claims)
	if !ok {
		return "", errors.New("no claims found in context")
	}
	return claims.UserID, nil
}

// RequireRole middleware đảm bảo người dùng có role cụ thể
func RequireRole(role string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value("claims").(*Claims)
			if !ok {
				http.Error(w, "Unauthorized - no valid claims", http.StatusUnauthorized)
				return
			}

			if claims.Role != role && claims.Role != "admin" {
				http.Error(w, "Forbidden - insufficient role", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
