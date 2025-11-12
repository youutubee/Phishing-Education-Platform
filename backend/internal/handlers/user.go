package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"seap/internal/auth"
	"seap/internal/middleware"
)

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value(middleware.UserIDKey).(string)
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	usersCollection := h.DB.Collection("users")
	var user struct {
		ID    primitive.ObjectID `bson:"_id" json:"id"`
		Email string             `bson:"email" json:"email"`
		Role  string             `bson:"role" json:"role"`
	}

	err = usersCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		respondWithError(w, http.StatusNotFound, "User not found")
		return
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"id":    user.ID.Hex(),
		"email": user.Email,
		"role":  user.Role,
	})
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value(middleware.UserIDKey).(string)
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	usersCollection := h.DB.Collection("users")
	updateFields := bson.M{"updated_at": time.Now()}

	if req.Email != "" {
		// Check if email already exists
		var existingUser struct {
			ID primitive.ObjectID `bson:"_id"`
		}
		err := usersCollection.FindOne(ctx, bson.M{"email": req.Email, "_id": bson.M{"$ne": userID}}).Decode(&existingUser)
		if err == nil {
			respondWithError(w, http.StatusConflict, "Email already in use")
			return
		} else if err != mongo.ErrNoDocuments {
			respondWithError(w, http.StatusInternalServerError, "Database error")
			return
		}

		updateFields["email"] = req.Email
	}

	if req.Password != "" {
		hashedPassword, err := auth.HashPassword(req.Password)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to hash password")
			return
		}
		updateFields["password_hash"] = hashedPassword
	}

	_, err = usersCollection.UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{"$set": updateFields},
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update profile")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Profile updated successfully"})
}
