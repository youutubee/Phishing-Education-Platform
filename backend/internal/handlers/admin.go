package handlers

import (
	"net/http"
	"strconv"

	"seap/internal/middleware"
	"seap/internal/models"

	"github.com/gorilla/mux"
)

func (h *Handler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(
		"SELECT id, email, role, email_verified, created_at FROM users ORDER BY created_at DESC",
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer rows.Close()

	type UserResponse struct {
		ID            int    `json:"id"`
		Email         string `json:"email"`
		Role          string `json:"role"`
		EmailVerified bool   `json:"email_verified"`
		CreatedAt     string `json:"created_at"`
	}

	var users []UserResponse
	for rows.Next() {
		var user UserResponse
		err := rows.Scan(&user.ID, &user.Email, &user.Role, &user.EmailVerified, &user.CreatedAt)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to scan user")
			return
		}
		users = append(users, user)
	}

	respondWithJSON(w, http.StatusOK, users)
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	adminID := r.Context().Value(middleware.UserIDKey).(int)
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	if userID == adminID {
		respondWithError(w, http.StatusBadRequest, "Cannot delete yourself")
		return
	}

	// Check if user exists
	var exists bool
	err = h.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", userID).Scan(&exists)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}
	if !exists {
		respondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	// Delete user (cascade will handle related records)
	_, err = h.DB.Exec("DELETE FROM users WHERE id = $1", userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete user")
		return
	}

	// Log audit
	details := `{"deleted_user_id": ` + strconv.Itoa(userID) + `}`
	_, err = h.DB.Exec(
		"INSERT INTO audit_logs (admin_id, action, resource_type, resource_id, details) VALUES ($1, $2, $3, $4, $5)",
		adminID, "delete_user", "user", userID, details,
	)
	if err != nil {
		// Log error but don't fail the request
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "User deleted successfully"})
}

func (h *Handler) GetAuditLogs(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(
		`SELECT al.id, al.admin_id, al.action, al.resource_type, al.resource_id, al.details, al.created_at, u.email as admin_email
		 FROM audit_logs al JOIN users u ON al.admin_id = u.id ORDER BY al.created_at DESC LIMIT 100`,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer rows.Close()

	type AuditLogResponse struct {
		ID           int    `json:"id"`
		AdminID      int    `json:"admin_id"`
		AdminEmail   string `json:"admin_email"`
		Action       string `json:"action"`
		ResourceType string `json:"resource_type"`
		ResourceID   *int   `json:"resource_id"`
		Details      string `json:"details"`
		CreatedAt    string `json:"created_at"`
	}

	var logs []AuditLogResponse
	for rows.Next() {
		var log AuditLogResponse
		err := rows.Scan(&log.ID, &log.AdminID, &log.Action, &log.ResourceType, &log.ResourceID, &log.Details, &log.CreatedAt, &log.AdminEmail)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to scan audit log")
			return
		}
		logs = append(logs, log)
	}

	respondWithJSON(w, http.StatusOK, logs)
}

func (h *Handler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(`
		SELECT 
			u.id as user_id,
			u.email,
			COUNT(DISTINCT c.id) as total_campaigns,
			COUNT(DISTINCT CASE WHEN e.event_type = 'link_opened' OR e.event_type = 'clicked' THEN e.id END) as total_clicks,
			COUNT(DISTINCT CASE WHEN e.event_type = 'awareness_viewed' THEN e.id END) as total_conversions,
			COUNT(DISTINCT CASE WHEN c.status = 'rejected' THEN c.id END) as rejected_count
		FROM users u
		LEFT JOIN campaigns c ON u.id = c.user_id
		LEFT JOIN events e ON c.id = e.campaign_id
		WHERE u.role = 'user'
		GROUP BY u.id, u.email
		ORDER BY total_clicks DESC, total_conversions DESC, rejected_count ASC
		LIMIT 50
	`)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer rows.Close()

	var leaderboard []models.LeaderboardEntry
	for rows.Next() {
		var entry models.LeaderboardEntry
		err := rows.Scan(&entry.UserID, &entry.Email, &entry.TotalCampaigns, &entry.TotalClicks, &entry.TotalConversions, &entry.RejectedCount)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to scan leaderboard entry")
			return
		}
		// Calculate score: clicks * 2 + conversions * 5 - rejections * 10
		entry.Score = entry.TotalClicks*2 + entry.TotalConversions*5 - entry.RejectedCount*10
		leaderboard = append(leaderboard, entry)
	}

	respondWithJSON(w, http.StatusOK, leaderboard)
}
