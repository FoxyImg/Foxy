package middleware

import (
	"context"
	"foxy/internal/db"
	"foxy/internal/env"
	"log"
	"net/http"
	"strings"
	"time"
)

func VerifyAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Verify auth middleware", r.Method, r.URL.Path)
		appId := r.PathValue("appId")
		if appId == "" {
			http.Error(w, "Missing App Id", http.StatusBadRequest)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization Header", http.StatusUnauthorized)
			return
		}

		key := strings.TrimSpace(strings.Replace(authHeader, "Bearer", "", 1))
		if key == "" {
			http.Error(w, "Missing Authorization Header", http.StatusUnauthorized)
			return
		}

		if env.FoxyEnvironment.Isolated {
			if key != *env.FoxyEnvironment.APIKey {
				http.Error(w, "Not Authorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
			return
		}

		verifiedSecret, err := db.RedisGet(appId + ":api-key")
		if err == nil && verifiedSecret != nil {
			if key != *verifiedSecret {
				http.Error(w, "Not Authorized", http.StatusUnauthorized)
				return
			}
		}

		conn, err := db.NewConnection()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return

		}
		defer conn.Release()

		var existingKey string
		res := conn.QueryRow(context.Background(), "SELECT admin_key FROM apps where id = $1", appId)
		_ = res.Scan(&existingKey)

		if existingKey != key {
			http.Error(w, "Not Authorized", http.StatusUnauthorized)
			return
		}

		_ = db.RedisSet(appId+":api-key", key, time.Hour*24)

		next.ServeHTTP(w, r)
	})
}
