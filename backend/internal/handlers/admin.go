package handlers

import (
	"context"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"seap/internal/middleware"
	"seap/internal/models"

	"github.com/gorilla/mux"
)

func (h *Handler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	usersCollection := h.DB.Collection("users")
	cursor, err := usersCollection.Find(
		ctx,
		bson.M{},
		options.Find().SetSort(bson.M{"created_at": -1}),
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer cursor.Close(ctx)

	type UserResponse struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		Role          string `json:"role"`
		EmailVerified bool   `json:"email_verified"`
		CreatedAt     string `json:"created_at"`
	}

	var users []UserResponse
	for cursor.Next(ctx) {
		var user models.User
		if err := cursor.Decode(&user); err != nil {
			continue
		}
		users = append(users, UserResponse{
			ID:            user.ID.Hex(),
			Email:         user.Email,
			Role:          user.Role,
			EmailVerified: user.EmailVerified,
			CreatedAt:     user.CreatedAt.Format(time.RFC3339),
		})
	}

	respondWithJSON(w, http.StatusOK, users)
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	adminIDStr := r.Context().Value(middleware.UserIDKey).(string)
	adminID, err := primitive.ObjectIDFromHex(adminIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid admin ID")
		return
	}

	vars := mux.Vars(r)
	userID, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	if userID == adminID {
		respondWithError(w, http.StatusBadRequest, "Cannot delete yourself")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if user exists
	usersCollection := h.DB.Collection("users")
	var user models.User
	err = usersCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		respondWithError(w, http.StatusNotFound, "User not found")
		return
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	// Delete user (MongoDB will handle related records if we set up proper cleanup)
	_, err = usersCollection.DeleteOne(ctx, bson.M{"_id": userID})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete user")
		return
	}

	// Also delete related campaigns and events
	campaignsCollection := h.DB.Collection("campaigns")
	campaignCursor, _ := campaignsCollection.Find(ctx, bson.M{"user_id": userID})
	var campaignIDs []primitive.ObjectID
	for campaignCursor.Next(ctx) {
		var campaign models.Campaign
		if err := campaignCursor.Decode(&campaign); err == nil {
			campaignIDs = append(campaignIDs, campaign.ID)
		}
	}
	campaignCursor.Close(ctx)

	if len(campaignIDs) > 0 {
		campaignsCollection.DeleteMany(ctx, bson.M{"user_id": userID})
		eventsCollection := h.DB.Collection("events")
		eventsCollection.DeleteMany(ctx, bson.M{"campaign_id": bson.M{"$in": campaignIDs}})
	}

	// Log audit
	details := `{"deleted_user_id": "` + userID.Hex() + `"}`
	auditLog := models.AuditLog{
		ID:           primitive.NewObjectID(),
		AdminID:      adminID,
		Action:       "delete_user",
		ResourceType: "user",
		ResourceID:   &userID,
		Details:      details,
		CreatedAt:    time.Now(),
	}
	auditLogsCollection := h.DB.Collection("audit_logs")
	_, err = auditLogsCollection.InsertOne(ctx, auditLog)
	if err != nil {
		// Log error but don't fail the request
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "User deleted successfully"})
}

func (h *Handler) GetAuditLogs(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	auditLogsCollection := h.DB.Collection("audit_logs")
	cursor, err := auditLogsCollection.Find(
		ctx,
		bson.M{},
		options.Find().SetSort(bson.M{"created_at": -1}).SetLimit(100),
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer cursor.Close(ctx)

	type AuditLogResponse struct {
		ID           string  `json:"id"`
		AdminID      string  `json:"admin_id"`
		AdminEmail   string  `json:"admin_email"`
		Action       string  `json:"action"`
		ResourceType string  `json:"resource_type"`
		ResourceID   *string `json:"resource_id"`
		Details      string  `json:"details"`
		CreatedAt    string  `json:"created_at"`
	}

	var logs []AuditLogResponse
	usersCollection := h.DB.Collection("users")
	for cursor.Next(ctx) {
		var log models.AuditLog
		if err := cursor.Decode(&log); err != nil {
			continue
		}

		var adminEmail string
		var adminUser models.User
		if err := usersCollection.FindOne(ctx, bson.M{"_id": log.AdminID}).Decode(&adminUser); err == nil {
			adminEmail = adminUser.Email
		}

		var resourceIDStr *string
		if log.ResourceID != nil {
			idStr := log.ResourceID.Hex()
			resourceIDStr = &idStr
		}

		logs = append(logs, AuditLogResponse{
			ID:           log.ID.Hex(),
			AdminID:      log.AdminID.Hex(),
			AdminEmail:   adminEmail,
			Action:       log.Action,
			ResourceType: log.ResourceType,
			ResourceID:   resourceIDStr,
			Details:      log.Details,
			CreatedAt:    log.CreatedAt.Format(time.RFC3339),
		})
	}

	respondWithJSON(w, http.StatusOK, logs)
}

func (h *Handler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get all users with role 'user'
	usersCollection := h.DB.Collection("users")
	userCursor, err := usersCollection.Find(ctx, bson.M{"role": "user"})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer userCursor.Close(ctx)

	var leaderboard []models.LeaderboardEntry
	campaignsCollection := h.DB.Collection("campaigns")
	eventsCollection := h.DB.Collection("events")

	for userCursor.Next(ctx) {
		var user models.User
		if err := userCursor.Decode(&user); err != nil {
			continue
		}

		// Count campaigns
		campaignCount, _ := campaignsCollection.CountDocuments(ctx, bson.M{"user_id": user.ID})

		// Get campaign IDs
		campaignCursor, _ := campaignsCollection.Find(ctx, bson.M{"user_id": user.ID})
		var campaignIDs []primitive.ObjectID
		for campaignCursor.Next(ctx) {
			var campaign models.Campaign
			if err := campaignCursor.Decode(&campaign); err == nil {
				campaignIDs = append(campaignIDs, campaign.ID)
			}
		}
		campaignCursor.Close(ctx)

		// Count clicks (link_opened or clicked)
		clickCount, _ := eventsCollection.CountDocuments(ctx, bson.M{
			"campaign_id": bson.M{"$in": campaignIDs},
			"event_type":  bson.M{"$in": []string{"link_opened", "clicked"}},
		})

		// Count conversions (awareness_viewed)
		conversionCount, _ := eventsCollection.CountDocuments(ctx, bson.M{
			"campaign_id": bson.M{"$in": campaignIDs},
			"event_type":  "awareness_viewed",
		})

		// Count rejected campaigns
		rejectedCount, _ := campaignsCollection.CountDocuments(ctx, bson.M{
			"user_id": user.ID,
			"status":  "rejected",
		})

		// Calculate score: clicks * 2 + conversions * 5 - rejections * 10
		score := int(clickCount)*2 + int(conversionCount)*5 - int(rejectedCount)*10

		leaderboard = append(leaderboard, models.LeaderboardEntry{
			UserID:           user.ID,
			Email:            user.Email,
			TotalCampaigns:   int(campaignCount),
			TotalClicks:      int(clickCount),
			TotalConversions: int(conversionCount),
			RejectedCount:    int(rejectedCount),
			Score:            score,
		})
	}

	// Sort by score (descending)
	for i := 0; i < len(leaderboard); i++ {
		for j := i + 1; j < len(leaderboard); j++ {
			if leaderboard[i].Score < leaderboard[j].Score {
				leaderboard[i], leaderboard[j] = leaderboard[j], leaderboard[i]
			}
		}
	}

	// Limit to top 50
	if len(leaderboard) > 50 {
		leaderboard = leaderboard[:50]
	}

	respondWithJSON(w, http.StatusOK, leaderboard)
}
