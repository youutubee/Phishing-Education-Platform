package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"seap/internal/middleware"
	"seap/internal/models"
	"seap/internal/utils"

	"github.com/gorilla/mux"
)

func (h *Handler) CreateCampaign(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	var req models.CampaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Title == "" || req.EmailText == "" {
		respondWithError(w, http.StatusBadRequest, "Title and email text are required")
		return
	}

	// Generate unique tracking token
	token, err := utils.GenerateToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	var campaignID int
	err = h.DB.QueryRow(
		`INSERT INTO campaigns (user_id, title, description, email_text, landing_page_url, tracking_token, status, expiry_date)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		userID, req.Title, req.Description, req.EmailText, req.LandingPageURL, token, "pending", req.ExpiryDate,
	).Scan(&campaignID)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create campaign")
		return
	}

	var campaign models.Campaign
	err = h.DB.QueryRow(
		`SELECT id, user_id, title, description, email_text, landing_page_url, tracking_token, status, expiry_date, admin_comment, created_at, updated_at
		 FROM campaigns WHERE id = $1`,
		campaignID,
	).Scan(&campaign.ID, &campaign.UserID, &campaign.Title, &campaign.Description, &campaign.EmailText,
		&campaign.LandingPageURL, &campaign.TrackingToken, &campaign.Status, &campaign.ExpiryDate,
		&campaign.AdminComment, &campaign.CreatedAt, &campaign.UpdatedAt)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve campaign")
		return
	}

	respondWithJSON(w, http.StatusCreated, campaign)
}

func (h *Handler) GetUserCampaigns(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	rows, err := h.DB.Query(
		`SELECT id, user_id, title, description, email_text, landing_page_url, tracking_token, status, expiry_date, admin_comment, created_at, updated_at
		 FROM campaigns WHERE user_id = $1 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer rows.Close()

	var campaigns []models.Campaign
	for rows.Next() {
		var campaign models.Campaign
		err := rows.Scan(&campaign.ID, &campaign.UserID, &campaign.Title, &campaign.Description, &campaign.EmailText,
			&campaign.LandingPageURL, &campaign.TrackingToken, &campaign.Status, &campaign.ExpiryDate,
			&campaign.AdminComment, &campaign.CreatedAt, &campaign.UpdatedAt)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to scan campaign")
			return
		}
		campaigns = append(campaigns, campaign)
	}

	respondWithJSON(w, http.StatusOK, campaigns)
}

func (h *Handler) GetCampaign(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)
	vars := mux.Vars(r)
	campaignID, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	var campaign models.Campaign
	err = h.DB.QueryRow(
		`SELECT id, user_id, title, description, email_text, landing_page_url, tracking_token, status, expiry_date, admin_comment, created_at, updated_at
		 FROM campaigns WHERE id = $1 AND user_id = $2`,
		campaignID, userID,
	).Scan(&campaign.ID, &campaign.UserID, &campaign.Title, &campaign.Description, &campaign.EmailText,
		&campaign.LandingPageURL, &campaign.TrackingToken, &campaign.Status, &campaign.ExpiryDate,
		&campaign.AdminComment, &campaign.CreatedAt, &campaign.UpdatedAt)

	if err == sql.ErrNoRows {
		respondWithError(w, http.StatusNotFound, "Campaign not found")
		return
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	respondWithJSON(w, http.StatusOK, campaign)
}

func (h *Handler) UpdateCampaign(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)
	vars := mux.Vars(r)
	campaignID, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Check if campaign belongs to user
	var ownerID int
	err = h.DB.QueryRow("SELECT user_id FROM campaigns WHERE id = $1", campaignID).Scan(&ownerID)
	if err == sql.ErrNoRows {
		respondWithError(w, http.StatusNotFound, "Campaign not found")
		return
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	if ownerID != userID {
		respondWithError(w, http.StatusForbidden, "Not authorized to update this campaign")
		return
	}

	var req models.CampaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	_, err = h.DB.Exec(
		`UPDATE campaigns SET title = $1, description = $2, email_text = $3, landing_page_url = $4, expiry_date = $5, updated_at = CURRENT_TIMESTAMP
		 WHERE id = $6`,
		req.Title, req.Description, req.EmailText, req.LandingPageURL, req.ExpiryDate, campaignID,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update campaign")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Campaign updated successfully"})
}

func (h *Handler) DeleteCampaign(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)
	vars := mux.Vars(r)
	campaignID, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Check if campaign belongs to user
	var ownerID int
	err = h.DB.QueryRow("SELECT user_id FROM campaigns WHERE id = $1", campaignID).Scan(&ownerID)
	if err == sql.ErrNoRows {
		respondWithError(w, http.StatusNotFound, "Campaign not found")
		return
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	if ownerID != userID {
		respondWithError(w, http.StatusForbidden, "Not authorized to delete this campaign")
		return
	}

	_, err = h.DB.Exec("DELETE FROM campaigns WHERE id = $1", campaignID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete campaign")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Campaign deleted successfully"})
}

func (h *Handler) GetAllCampaigns(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(
		`SELECT c.id, c.user_id, c.title, c.description, c.email_text, c.landing_page_url, c.tracking_token, c.status, c.expiry_date, c.admin_comment, c.created_at, c.updated_at, u.email
		 FROM campaigns c JOIN users u ON c.user_id = u.id ORDER BY c.created_at DESC`,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer rows.Close()

	type CampaignWithUser struct {
		models.Campaign
		UserEmail string `json:"user_email"`
	}

	var campaigns []CampaignWithUser
	for rows.Next() {
		var campaign CampaignWithUser
		err := rows.Scan(&campaign.ID, &campaign.UserID, &campaign.Title, &campaign.Description, &campaign.EmailText,
			&campaign.LandingPageURL, &campaign.TrackingToken, &campaign.Status, &campaign.ExpiryDate,
			&campaign.AdminComment, &campaign.CreatedAt, &campaign.UpdatedAt, &campaign.UserEmail)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to scan campaign")
			return
		}
		campaigns = append(campaigns, campaign)
	}

	respondWithJSON(w, http.StatusOK, campaigns)
}

func (h *Handler) ApproveCampaign(w http.ResponseWriter, r *http.Request) {
	adminID := r.Context().Value(middleware.UserIDKey).(int)
	vars := mux.Vars(r)
	campaignID, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	var req models.CampaignApprovalRequest
	json.NewDecoder(r.Body).Decode(&req)

	// Update campaign status
	_, err = h.DB.Exec(
		"UPDATE campaigns SET status = 'approved', admin_comment = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2",
		req.Comment, campaignID,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to approve campaign")
		return
	}

	// Log audit
	details := `{"comment": "` + req.Comment + `"}`
	_, err = h.DB.Exec(
		"INSERT INTO audit_logs (admin_id, action, resource_type, resource_id, details) VALUES ($1, $2, $3, $4, $5)",
		adminID, "approve_campaign", "campaign", campaignID, details,
	)
	if err != nil {
		// Log error but don't fail the request
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Campaign approved successfully"})
}

func (h *Handler) RejectCampaign(w http.ResponseWriter, r *http.Request) {
	adminID := r.Context().Value(middleware.UserIDKey).(int)
	vars := mux.Vars(r)
	campaignID, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	var req models.CampaignApprovalRequest
	json.NewDecoder(r.Body).Decode(&req)

	if req.Comment == "" {
		respondWithError(w, http.StatusBadRequest, "Comment is required for rejection")
		return
	}

	// Update campaign status
	_, err = h.DB.Exec(
		"UPDATE campaigns SET status = 'rejected', admin_comment = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2",
		req.Comment, campaignID,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to reject campaign")
		return
	}

	// Log audit
	details := `{"comment": "` + req.Comment + `"}`
	_, err = h.DB.Exec(
		"INSERT INTO audit_logs (admin_id, action, resource_type, resource_id, details) VALUES ($1, $2, $3, $4, $5)",
		adminID, "reject_campaign", "campaign", campaignID, details,
	)
	if err != nil {
		// Log error but don't fail the request
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Campaign rejected successfully"})
}
