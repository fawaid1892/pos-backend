package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"pos-multi-branch/backend/internal/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type contextKey string

const (
	UserIDKey   contextKey = "user_id"
	UserRoleKey contextKey = "user_role"
	RoleIDKey   contextKey = "role_id"
	BranchIDKey contextKey = "branch_id"
)

var jwtSecret []byte

func InitJWT(cfg *config.Config) {
	jwtSecret = []byte(cfg.JWTSecret)
}

func GenerateToken(userID uuid.UUID, role string, roleID *uuid.UUID, branchID *uuid.UUID, expiryHours int) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"role":    role,
		"exp":     time.Now().Add(time.Duration(expiryHours) * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}
	if roleID != nil {
		claims["role_id"] = roleID.String()
	}
	if branchID != nil {
		claims["branch_id"] = branchID.String()
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			http.Error(w, `{"error":"invalid authorization format"}`, http.StatusUnauthorized)
			return
		}

		tokenStr := parts[1]
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, `{"error":"invalid token claims"}`, http.StatusUnauthorized)
			return
		}

		userIDStr, _ := claims["user_id"].(string)
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			http.Error(w, `{"error":"invalid user_id in token"}`, http.StatusUnauthorized)
			return
		}

		role, _ := claims["role"].(string)
		roleIDStr, _ := claims["role_id"].(string)
		branchIDStr, _ := claims["branch_id"].(string)

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		ctx = context.WithValue(ctx, UserRoleKey, role)
		if roleIDStr != "" {
			if rid, err := uuid.Parse(roleIDStr); err == nil {
				ctx = context.WithValue(ctx, RoleIDKey, rid)
			}
		}
		if branchIDStr != "" {
			if bid, err := uuid.Parse(branchIDStr); err == nil {
				ctx = context.WithValue(ctx, BranchIDKey, bid)
			}
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserID(ctx context.Context) uuid.UUID {
	v, _ := ctx.Value(UserIDKey).(uuid.UUID)
	return v
}

func GetUserRole(ctx context.Context) string {
	v, _ := ctx.Value(UserRoleKey).(string)
	return v
}

func GetUserRoleID(ctx context.Context) *uuid.UUID {
	v, ok := ctx.Value(RoleIDKey).(uuid.UUID)
	if !ok {
		return nil
	}
	return &v
}

func GetBranchID(ctx context.Context) *uuid.UUID {
	v, ok := ctx.Value(BranchIDKey).(uuid.UUID)
	if !ok {
		return nil
	}
	return &v
}

// RequireRole returns a middleware that checks if the authenticated user has one
// of the allowed roles. Must be used after AuthMiddleware.
func RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := GetUserRole(r.Context())
			if role == "" {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			for _, allowed := range allowedRoles {
				if role == allowed {
					next.ServeHTTP(w, r)
					return
				}
			}
			http.Error(w, `{"error":"forbidden: insufficient permissions"}`, http.StatusForbidden)
		})
	}
}

// RequirePermission returns a middleware that checks if the authenticated user's
// role has a specific permission. Must be used after AuthMiddleware.
func RequirePermission(permissionName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			roleID := GetUserRoleID(r.Context())
			if roleID == nil {
				http.Error(w, `{"error":"forbidden: no role assigned"}`, http.StatusForbidden)
				return
			}
			has, err := RoleHasPermission(*roleID, permissionName)
			if err != nil {
				http.Error(w, `{"error":"internal error checking permissions"}`, http.StatusInternalServerError)
				return
			}
			if !has {
				http.Error(w, `{"error":"forbidden: insufficient permissions"}`, http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RoleHasPermission checks if a role has a specific permission by name.
// Defined here in middleware to avoid circular imports with repository.
var RoleHasPermission func(roleID uuid.UUID, permissionName string) (bool, error)
