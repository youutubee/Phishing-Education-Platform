package main

import (
	"log"
	"net/http"
	"os"

	"seap/internal/database"
	"seap/internal/handlers"
	"seap/internal/middleware"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Initialize database
	db, err := database.InitDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Initialize handlers
	h := handlers.NewHandler(db)

	// Setup router
	r := mux.NewRouter()

	// CORS middleware
	r.Use(middleware.CORS)

	// Public routes
	r.HandleFunc("/api/health", handlers.HealthCheck).Methods("GET")
	r.HandleFunc("/api/auth/register", h.Register).Methods("POST")
	r.HandleFunc("/api/auth/login", h.Login).Methods("POST")
	r.HandleFunc("/api/auth/verify-otp", h.VerifyOTP).Methods("POST")
	r.HandleFunc("/api/auth/resend-otp", h.ResendOTP).Methods("POST")
	r.HandleFunc("/api/simulate/{token}", h.SimulateLanding).Methods("GET")
	r.HandleFunc("/api/simulate/{token}/submit", h.SimulateSubmit).Methods("POST")
	r.HandleFunc("/api/awareness/{token}", h.AwarenessPage).Methods("GET")

	// Protected routes (require authentication)
	api := r.PathPrefix("/api").Subrouter()
	api.Use(middleware.AuthMiddleware)

	// User routes
	api.HandleFunc("/user/profile", h.GetProfile).Methods("GET")
	api.HandleFunc("/user/profile", h.UpdateProfile).Methods("PUT")
	api.HandleFunc("/user/campaigns", h.CreateCampaign).Methods("POST")
	api.HandleFunc("/user/campaigns", h.GetUserCampaigns).Methods("GET")
	api.HandleFunc("/user/campaigns/{id}", h.GetCampaign).Methods("GET")
	api.HandleFunc("/user/campaigns/{id}", h.UpdateCampaign).Methods("PUT")
	api.HandleFunc("/user/campaigns/{id}", h.DeleteCampaign).Methods("DELETE")
	api.HandleFunc("/user/analytics", h.GetUserAnalytics).Methods("GET")

	// Admin routes
	admin := api.PathPrefix("/admin").Subrouter()
	admin.Use(middleware.AdminMiddleware)

	admin.HandleFunc("/campaigns", h.GetAllCampaigns).Methods("GET")
	admin.HandleFunc("/campaigns/{id}/approve", h.ApproveCampaign).Methods("POST")
	admin.HandleFunc("/campaigns/{id}/reject", h.RejectCampaign).Methods("POST")
	admin.HandleFunc("/users", h.GetAllUsers).Methods("GET")
	admin.HandleFunc("/users/{id}", h.DeleteUser).Methods("DELETE")
	admin.HandleFunc("/audit-logs", h.GetAuditLogs).Methods("GET")
	admin.HandleFunc("/leaderboard", h.GetLeaderboard).Methods("GET")
	admin.HandleFunc("/analytics", h.GetAdminAnalytics).Methods("GET")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

