package main

import (
	"net/http"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-API-KEY")
		w.Header().Set("Content-Type", "application/json")

		if key == "" {
			w.WriteHeader(http.StatusForbidden)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
