package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/OrioXZ/7solutions-backend-challenge/internal/security"
)

type HealthChecker interface {
	Ping(ctx context.Context) error
}

type healthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
}

func NewRouter(
	healthChecker HealthChecker,
	authenticationService AuthenticationService,
	userService UserManagementService,
	tokenValidator security.TokenValidator,
) http.Handler {
	mux := http.NewServeMux()
	authHandler := NewAuthHandler(authenticationService)
	userHandler := NewUserHandler(userService)
	authenticate := AuthenticationMiddleware(tokenValidator)

	mux.HandleFunc("GET /health", handleHealth(healthChecker))
	mux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)

	mux.Handle("POST /api/v1/users", authenticate(http.HandlerFunc(userHandler.Create)))
	mux.Handle("GET /api/v1/users", authenticate(http.HandlerFunc(userHandler.List)))
	mux.Handle("GET /api/v1/users/{id}", authenticate(http.HandlerFunc(userHandler.GetByID)))
	mux.Handle("PATCH /api/v1/users/{id}", authenticate(http.HandlerFunc(userHandler.Update)))
	mux.Handle("DELETE /api/v1/users/{id}", authenticate(http.HandlerFunc(userHandler.Delete)))

	return mux
}

func handleHealth(healthChecker HealthChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		response := healthResponse{
			Status:   "ok",
			Database: "connected",
		}
		statusCode := http.StatusOK

		if err := healthChecker.Ping(ctx); err != nil {
			response.Status = "unavailable"
			response.Database = "disconnected"
			statusCode = http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		_ = json.NewEncoder(w).Encode(response)
	}
}
