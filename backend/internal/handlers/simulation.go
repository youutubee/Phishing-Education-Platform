package handlers

import (
	"context"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/gorilla/mux"
)

func (h *Handler) SimulateLanding(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	token := vars["token"]

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get campaign by token
	campaignsCollection := h.DB.Collection("campaigns")
	var campaign struct {
		ID             primitive.ObjectID `bson:"_id" json:"id"`
		Title          string             `bson:"title" json:"title"`
		LandingPageURL string             `bson:"landing_page_url" json:"landing_page_url"`
		Status         string             `bson:"status" json:"status"`
		ExpiryDate     *time.Time         `bson:"expiry_date,omitempty" json:"expiry_date"`
	}

	err := campaignsCollection.FindOne(ctx, bson.M{"tracking_token": token}).Decode(&campaign)
	if err == mongo.ErrNoDocuments {
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

	// Expiry date check is optional - campaigns work indefinitely unless explicitly expired
	// Uncomment the following block if you want to enforce expiry dates:
	/*
		if campaign.ExpiryDate != nil && !campaign.ExpiryDate.IsZero() {
			now := time.Now()
			if now.After(*campaign.ExpiryDate) {
				respondWithError(w, http.StatusGone, "Campaign has expired")
				return
			}
		}
	*/

	// Log link opened event (with deduplication - prevent duplicate events within 5 seconds)
	ipAddress := r.RemoteAddr
	userAgent := r.UserAgent()
	eventsCollection := h.DB.Collection("events")

	// Check if a similar event was logged recently (within last 5 seconds)
	fiveSecondsAgo := time.Now().Add(-5 * time.Second)
	recentEventCount, _ := eventsCollection.CountDocuments(ctx, bson.M{
		"campaign_id": campaign.ID,
		"event_type":  "link_opened",
		"ip_address":  ipAddress,
		"created_at":  bson.M{"$gte": fiveSecondsAgo},
	})

	// Only log if no recent event exists (prevents duplicate logging from React StrictMode or double-clicks)
	if recentEventCount == 0 {
		event := map[string]interface{}{
			"campaign_id": campaign.ID,
			"event_type":  "link_opened",
			"ip_address":  ipAddress,
			"user_agent":  userAgent,
			"created_at":  time.Now(),
		}
		_, err = eventsCollection.InsertOne(ctx, event)
		if err != nil {
			// Log error but continue
		}
	}

	// Return landing page data
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"campaign_id": campaign.ID.Hex(),
		"title":       campaign.Title,
		"landing_url": campaign.LandingPageURL,
		"token":       token,
	})
}

func (h *Handler) SimulateSubmit(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	token := vars["token"]

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get campaign by token
	campaignsCollection := h.DB.Collection("campaigns")
	var campaign struct {
		ID     primitive.ObjectID `bson:"_id"`
		Status string             `bson:"status"`
	}

	err := campaignsCollection.FindOne(ctx, bson.M{"tracking_token": token}).Decode(&campaign)
	if err == mongo.ErrNoDocuments {
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

	// Log form submission event
	ipAddress := r.RemoteAddr
	userAgent := r.UserAgent()
	event := map[string]interface{}{
		"campaign_id": campaign.ID,
		"event_type":  "form_submitted",
		"ip_address":  ipAddress,
		"user_agent":  userAgent,
		"created_at":  time.Now(),
	}
	eventsCollection := h.DB.Collection("events")
	_, err = eventsCollection.InsertOne(ctx, event)
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get campaign by token
	campaignsCollection := h.DB.Collection("campaigns")
	var campaign struct {
		ID primitive.ObjectID `bson:"_id"`
	}

	err := campaignsCollection.FindOne(ctx, bson.M{"tracking_token": token}).Decode(&campaign)
	if err == mongo.ErrNoDocuments {
		respondWithError(w, http.StatusNotFound, "Campaign not found")
		return
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	// Log awareness page view
	ipAddress := r.RemoteAddr
	userAgent := r.UserAgent()
	event := map[string]interface{}{
		"campaign_id": campaign.ID,
		"event_type":  "awareness_viewed",
		"ip_address":  ipAddress,
		"user_agent":  userAgent,
		"created_at":  time.Now(),
	}
	eventsCollection := h.DB.Collection("events")
	_, err = eventsCollection.InsertOne(ctx, event)
	if err != nil {
		// Log error but continue
	}

	// Return awareness content
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"campaign_id": campaign.ID.Hex(),
		"message":     "This was a simulated phishing attempt",
		"content": map[string]string{
			"title":       "You've Been Phished! (Simulated)",
			"description": "This was a safe, educational simulation designed to teach you about phishing attacks.",
			"tips":        "Always verify sender emails, check URLs carefully, and never enter credentials on suspicious pages.",
		},
	})
}
