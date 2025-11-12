package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"seap/internal/auth"
	"seap/internal/middleware"
)

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	var user struct {
		ID    int    `json:"id"`
		Email string `json:"email"`
		Role  string `json:"role"`
	}

	err := h.DB.QueryRow(
		"SELECT id, email, role FROM users WHERE id = $1",
		userID,
	).Scan(&user.ID, &user.Email, &user.Role)

	if err == sql.ErrNoRows {
		respondWithError(w, http.StatusNotFound, "User not found")
		return
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	respondWithJSON(w, http.StatusOK, user)
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email != "" {
		// Check if email already exists
		var existingID int
		err := h.DB.QueryRow("SELECT id FROM users WHERE email = $1 AND id != $2", req.Email, userID).Scan(&existingID)
		if err == nil {
			respondWithError(w, http.StatusConflict, "Email already in use")
			return
		} else if err != sql.ErrNoRows {
			respondWithError(w, http.StatusInternalServerError, "Database error")
			return
		}

		_, err = h.DB.Exec("UPDATE users SET email = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2", req.Email, userID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to update email")
			return
		}
	}

	if req.Password != "" {
		hashedPassword, err := auth.HashPassword(req.Password)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to hash password")
			return
		}

		_, err = h.DB.Exec("UPDATE users SET password_hash = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2", hashedPassword, userID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to update password")
			return
		}
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Profile updated successfully"})
}

