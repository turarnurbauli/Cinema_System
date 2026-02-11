package middleware

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// RequireRoleForMethods оборачивает handler и требует наличие одной из ролей
// для указанных HTTP-методов. Для остальных методов доступ свободный.
func RequireRoleForMethods(next http.Handler, jwtSecret string, methods map[string][]string) http.Handler {
	secret := []byte(jwtSecret)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowedRoles, ok := methods[r.Method]
		if !ok {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, `{"error":"missing or invalid Authorization header"}`, http.StatusUnauthorized)
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return secret, nil
		})
		if err != nil || !token.Valid {
			http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, `{"error":"invalid token claims"}`, http.StatusUnauthorized)
			return
		}

		roleVal, ok := claims["role"].(string)
		if !ok {
			http.Error(w, `{"error":"invalid token role"}`, http.StatusUnauthorized)
			return
		}

		for _, rname := range allowedRoles {
			if rname == roleVal {
				next.ServeHTTP(w, r)
				return
			}
		}

		http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
	})
}

