package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func (h *Handler) SimulateLanding(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	token := vars["token"]

	// Get campaign by token
	var campaign struct {
		ID            int
		Title         string
		LandingPageURL string
		Status        string
		ExpiryDate    *time.Time
	}

	err := h.DB.QueryRow(
		"SELECT id, title, landing_page_url, status, expiry_date FROM campaigns WHERE tracking_token = $1",
		token,
	).Scan(&campaign.ID, &campaign.Title, &campaign.LandingPageURL, &campaign.Status, &campaign.ExpiryDate)

	if err == sql.ErrNoRows {
		respondWithError(w, http.StatusNotFound, "Campaign not found")
		return
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	if campaign.Status != "approved" {
		respondWithError(w, http.StatusForbidden, "Campaign not approved")
		return
	}

	if campaign.ExpiryDate != nil && time.Now().After(*campaign.ExpiryDate) {
		respondWithError(w, http.StatusGone, "Campaign has expired")
		return
	}

	// Log link opened event
	ipAddress := r.RemoteAddr
	userAgent := r.UserAgent()
	_, err = h.DB.Exec(
		"INSERT INTO events (campaign_id, event_type, ip_address, user_agent) VALUES ($1, $2, $3, $4)",
		campaign.ID, "link_opened", ipAddress, userAgent,
	)
	if err != nil {
		// Log error but continue
	}

	// Return landing page data
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"campaign_id": campaign.ID,
		"title":       campaign.Title,
		"landing_url": campaign.LandingPageURL,
		"token":       token,
	})
}

func (h *Handler) SimulateSubmit(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	token := vars["token"]

	// Get campaign by token
	var campaignID int
	var status string
	err := h.DB.QueryRow(
		"SELECT id, status FROM campaigns WHERE tracking_token = $1",
		token,
	).Scan(&campaignID, &status)

	if err == sql.ErrNoRows {
		respondWithError(w, http.StatusNotFound, "Campaign not found")
		return
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	if status != "approved" {
		respondWithError(w, http.StatusForbidden, "Campaign not approved")
		return
	}

	// Log form submission event
	ipAddress := r.RemoteAddr
	userAgent := r.UserAgent()
	_, err = h.DB.Exec(
		"INSERT INTO events (campaign_id, event_type, ip_address, user_agent) VALUES ($1, $2, $3, $4)",
		campaignID, "form_submitted", ipAddress, userAgent,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to log event")
		return
	}

	// Redirect to awareness page
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"redirect": "/api/awareness/" + token,
		"message":  "Form submitted (simulated)",
	})
}

func (h *Handler) AwarenessPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	token := vars["token"]

	// Get campaign by token
	var campaignID int
	err := h.DB.QueryRow(
		"SELECT id FROM campaigns WHERE tracking_token = $1",
		token,
	).Scan(&campaignID)

	if err == sql.ErrNoRows {
		respondWithError(w, http.StatusNotFound, "Campaign not found")
		return
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	// Log awareness page view
	ipAddress := r.RemoteAddr
	userAgent := r.UserAgent()
	_, err = h.DB.Exec(
		"INSERT INTO events (campaign_id, event_type, ip_address, user_agent) VALUES ($1, $2, $3, $4)",
		campaignID, "awareness_viewed", ipAddress, userAgent,
	)
	if err != nil {
		// Log error but continue
	}

	// Return awareness content
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"campaign_id": campaignID,
		"message":     "This was a simulated phishing attempt",
		"content": map[string]string{
			"title":       "You've Been Phished! (Simulated)",
			"description": "This was a safe, educational simulation designed to teach you about phishing attacks.",
			"tips":        "Always verify sender emails, check URLs carefully, and never enter credentials on suspicious pages.",
		},
	})
}

