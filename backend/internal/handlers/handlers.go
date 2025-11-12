package handlers

import (
	"encoding/json"
	"net/http"

	"go.mongodb.org/mongo-driver/mongo"
)

type Handler struct {
	DB *mongo.Database
}

func NewHandler(db *mongo.Database) *Handler {
	return &Handler{DB: db}
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
}
