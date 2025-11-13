package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"seap/internal/middleware"
	"seap/internal/models"
	"seap/internal/utils"

	"github.com/gorilla/mux"
)

func (h *Handler) CreateCampaign(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value(middleware.UserIDKey).(string)
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	campaign := models.Campaign{
		ID:             primitive.NewObjectID(),
		UserID:         userID,
		Title:          req.Title,
		Description:    req.Description,
		EmailText:      req.EmailText,
		LandingPageURL: req.LandingPageURL,
		TrackingToken:  token,
		Status:         "pending",
		ExpiryDate:     req.ExpiryDate,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	campaignsCollection := h.DB.Collection("campaigns")
	_, err = campaignsCollection.InsertOne(ctx, campaign)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create campaign")
		return
	}

	respondWithJSON(w, http.StatusCreated, campaign)
}

func (h *Handler) GetUserCampaigns(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value(middleware.UserIDKey).(string)
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	campaignsCollection := h.DB.Collection("campaigns")
	cursor, err := campaignsCollection.Find(
		ctx,
		bson.M{"user_id": userID},
		options.Find().SetSort(bson.M{"created_at": -1}),
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer cursor.Close(ctx)

	var campaigns []models.Campaign
	if err = cursor.All(ctx, &campaigns); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve campaigns")
		return
	}

	respondWithJSON(w, http.StatusOK, campaigns)
}

func (h *Handler) GetCampaign(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value(middleware.UserIDKey).(string)
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	vars := mux.Vars(r)
	campaignID, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	campaignsCollection := h.DB.Collection("campaigns")
	var campaign models.Campaign
	err = campaignsCollection.FindOne(
		ctx,
		bson.M{"_id": campaignID, "user_id": userID},
	).Decode(&campaign)

	if err == mongo.ErrNoDocuments {
		respondWithError(w, http.StatusNotFound, "Campaign not found")
		return
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	respondWithJSON(w, http.StatusOK, campaign)
}

func (h *Handler) UpdateCampaign(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value(middleware.UserIDKey).(string)
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	vars := mux.Vars(r)
	campaignID, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if campaign belongs to user
	campaignsCollection := h.DB.Collection("campaigns")
	var campaign models.Campaign
	err = campaignsCollection.FindOne(ctx, bson.M{"_id": campaignID}).Decode(&campaign)
	if err == mongo.ErrNoDocuments {
		respondWithError(w, http.StatusNotFound, "Campaign not found")
		return
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	if campaign.UserID != userID {
		respondWithError(w, http.StatusForbidden, "Not authorized to update this campaign")
		return
	}

	var req models.CampaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	updateFields := bson.M{
		"title":            req.Title,
		"description":      req.Description,
		"email_text":       req.EmailText,
		"landing_page_url": req.LandingPageURL,
		"updated_at":       time.Now(),
	}

	// Handle expiry date - set if provided, unset if nil
	updateOp := bson.M{"$set": updateFields}
	if req.ExpiryDate != nil {
		updateFields["expiry_date"] = req.ExpiryDate
	} else {
		updateOp["$unset"] = bson.M{"expiry_date": ""}
	}

	_, err = campaignsCollection.UpdateOne(
		ctx,
		bson.M{"_id": campaignID},
		updateOp,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update campaign")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Campaign updated successfully"})
}

func (h *Handler) DeleteCampaign(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value(middleware.UserIDKey).(string)
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	vars := mux.Vars(r)
	campaignID, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if campaign belongs to user
	campaignsCollection := h.DB.Collection("campaigns")
	var campaign models.Campaign
	err = campaignsCollection.FindOne(ctx, bson.M{"_id": campaignID}).Decode(&campaign)
	if err == mongo.ErrNoDocuments {
		respondWithError(w, http.StatusNotFound, "Campaign not found")
		return
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	if campaign.UserID != userID {
		respondWithError(w, http.StatusForbidden, "Not authorized to delete this campaign")
		return
	}

	_, err = campaignsCollection.DeleteOne(ctx, bson.M{"_id": campaignID})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete campaign")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Campaign deleted successfully"})
}

func (h *Handler) GetAllCampaigns(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	campaignsCollection := h.DB.Collection("campaigns")
	cursor, err := campaignsCollection.Find(
		ctx,
		bson.M{},
		options.Find().SetSort(bson.M{"created_at": -1}),
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer cursor.Close(ctx)

	var campaigns []models.Campaign
	if err = cursor.All(ctx, &campaigns); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve campaigns")
		return
	}

	// Get user emails
	usersCollection := h.DB.Collection("users")
	type CampaignWithUser struct {
		models.Campaign
		UserEmail string `json:"user_email"`
	}

	var campaignsWithUser []CampaignWithUser
	for _, campaign := range campaigns {
		var user models.User
		err := usersCollection.FindOne(ctx, bson.M{"_id": campaign.UserID}).Decode(&user)
		if err == nil {
			campaignsWithUser = append(campaignsWithUser, CampaignWithUser{
				Campaign:  campaign,
				UserEmail: user.Email,
			})
		} else {
			campaignsWithUser = append(campaignsWithUser, CampaignWithUser{
				Campaign:  campaign,
				UserEmail: "",
			})
		}
	}

	respondWithJSON(w, http.StatusOK, campaignsWithUser)
}

func (h *Handler) ApproveCampaign(w http.ResponseWriter, r *http.Request) {
	adminIDStr := r.Context().Value(middleware.UserIDKey).(string)
	adminID, err := primitive.ObjectIDFromHex(adminIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid admin ID")
		return
	}

	vars := mux.Vars(r)
	campaignID, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	var req models.CampaignApprovalRequest
	json.NewDecoder(r.Body).Decode(&req)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Update campaign status
	campaignsCollection := h.DB.Collection("campaigns")
	var campaign models.Campaign
	err = campaignsCollection.FindOne(ctx, bson.M{"_id": campaignID}).Decode(&campaign)
	if err == mongo.ErrNoDocuments {
		respondWithError(w, http.StatusNotFound, "Campaign not found")
		return
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	_, err = campaignsCollection.UpdateOne(
		ctx,
		bson.M{"_id": campaignID},
		bson.M{"$set": bson.M{
			"status":        "approved",
			"admin_comment": req.Comment,
			"updated_at":    time.Now(),
		}},
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to approve campaign")
		return
	}

	// Log audit
	details := `{"comment": "` + req.Comment + `"}`
	auditLog := models.AuditLog{
		ID:           primitive.NewObjectID(),
		AdminID:      adminID,
		Action:       "approve_campaign",
		ResourceType: "campaign",
		ResourceID:   &campaignID,
		Details:      details,
		CreatedAt:    time.Now(),
	}
	auditLogsCollection := h.DB.Collection("audit_logs")
	_, err = auditLogsCollection.InsertOne(ctx, auditLog)
	if err != nil {
		// Log error but don't fail the request
	}

	// Send notification email to the campaign owner (the user who created the campaign)
	// NOTE: This is NOT the email from the phishing form - that's just simulated data
	campaign.Status = "approved"
	campaign.AdminComment = req.Comment
	usersCollection := h.DB.Collection("users")
	var owner models.User
	if err := usersCollection.FindOne(ctx, bson.M{"_id": campaign.UserID}).Decode(&owner); err == nil {
		frontendURL := strings.TrimSuffix(os.Getenv("APP_BASE_URL"), "/")
		if frontendURL == "" {
			frontendURL = "http://localhost:3000"
		}
		simulationLink := frontendURL + "/simulate/" + campaign.TrackingToken
		log.Printf("Sending approval email to campaign owner: %s", owner.Email)
		go utils.SendCampaignDecisionEmail(owner.Email, campaign.Title, "approved", req.Comment, simulationLink)
	} else {
		log.Printf("WARNING: Could not find campaign owner user ID: %s", campaign.UserID.Hex())
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Campaign approved successfully"})
}

func (h *Handler) RejectCampaign(w http.ResponseWriter, r *http.Request) {
	adminIDStr := r.Context().Value(middleware.UserIDKey).(string)
	adminID, err := primitive.ObjectIDFromHex(adminIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid admin ID")
		return
	}

	vars := mux.Vars(r)
	campaignID, err := primitive.ObjectIDFromHex(vars["id"])
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Update campaign status
	campaignsCollection := h.DB.Collection("campaigns")
	var campaign models.Campaign
	err = campaignsCollection.FindOne(ctx, bson.M{"_id": campaignID}).Decode(&campaign)
	if err == mongo.ErrNoDocuments {
		respondWithError(w, http.StatusNotFound, "Campaign not found")
		return
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	_, err = campaignsCollection.UpdateOne(
		ctx,
		bson.M{"_id": campaignID},
		bson.M{"$set": bson.M{
			"status":        "rejected",
			"admin_comment": req.Comment,
			"updated_at":    time.Now(),
		}},
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to reject campaign")
		return
	}

	// Log audit
	details := `{"comment": "` + req.Comment + `"}`
	auditLog := models.AuditLog{
		ID:           primitive.NewObjectID(),
		AdminID:      adminID,
		Action:       "reject_campaign",
		ResourceType: "campaign",
		ResourceID:   &campaignID,
		Details:      details,
		CreatedAt:    time.Now(),
	}
	auditLogsCollection := h.DB.Collection("audit_logs")
	_, err = auditLogsCollection.InsertOne(ctx, auditLog)
	if err != nil {
		// Log error but don't fail the request
	}

	// Send notification email to the campaign owner (the user who created the campaign)
	// NOTE: This is NOT the email from the phishing form - that's just simulated data
	campaign.Status = "rejected"
	campaign.AdminComment = req.Comment
	usersCollection := h.DB.Collection("users")
	var owner models.User
	if err := usersCollection.FindOne(ctx, bson.M{"_id": campaign.UserID}).Decode(&owner); err == nil {
		log.Printf("Sending rejection email to campaign owner: %s", owner.Email)
		go utils.SendCampaignDecisionEmail(owner.Email, campaign.Title, "rejected", req.Comment, "")
	} else {
		log.Printf("WARNING: Could not find campaign owner user ID: %s", campaign.UserID.Hex())
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Campaign rejected successfully"})
}

type ShareCampaignRequest struct {
	Email string `json:"email"`
}

func (h *Handler) ShareCampaign(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value(middleware.UserIDKey).(string)
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	vars := mux.Vars(r)
	campaignID, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	var req ShareCampaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" {
		respondWithError(w, http.StatusBadRequest, "Email address is required")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if campaign belongs to user
	campaignsCollection := h.DB.Collection("campaigns")
	var campaign models.Campaign
	err = campaignsCollection.FindOne(ctx, bson.M{"_id": campaignID, "user_id": userID}).Decode(&campaign)
	if err == mongo.ErrNoDocuments {
		respondWithError(w, http.StatusNotFound, "Campaign not found")
		return
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	if campaign.Status != "approved" {
		respondWithError(w, http.StatusBadRequest, "Only approved campaigns can be shared")
		return
	}

	// Generate simulation link
	frontendURL := strings.TrimSuffix(os.Getenv("APP_BASE_URL"), "/")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	simulationLink := frontendURL + "/simulate/" + campaign.TrackingToken

	// Check if Resend API is configured
	resendAPIKey := os.Getenv("RESEND_API_KEY")
	if resendAPIKey == "" {
		log.Printf("WARNING: RESEND_API_KEY not set; cannot send email")
		respondWithError(w, http.StatusServiceUnavailable, "Email service is not configured. Please contact administrator.")
		return
	}

	// Validate email format
	if !strings.Contains(req.Email, "@") {
		respondWithError(w, http.StatusBadRequest, "Invalid email address format")
		return
	}

	// Send email asynchronously but log errors
	emailSent := make(chan error, 1)
	go func() {
		err := utils.SendCampaignShareEmail(req.Email, campaign.Title, simulationLink)
		emailSent <- err
		if err != nil {
			log.Printf("ERROR: Failed to send campaign share email to %s: %v", req.Email, err)
		} else {
			log.Printf("Successfully sent campaign share email to %s for campaign: %s", req.Email, campaign.Title)
		}
	}()

	// Wait a short time to check if email sending started successfully
	select {
	case err := <-emailSent:
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to send email: %v", err))
			return
		}
	case <-time.After(100 * time.Millisecond):
		// Email sending started, return success
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Campaign link sent successfully",
		"email":   req.Email,
	})
}
