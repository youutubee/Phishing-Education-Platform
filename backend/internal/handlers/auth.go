package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"seap/internal/auth"
	"seap/internal/models"
)

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate email and password
	if req.Email == "" || req.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	// Validate role
	if req.Role != "user" && req.Role != "admin" {
		req.Role = "user"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if user exists
	usersCollection := h.DB.Collection("users")
	var existingUser models.User
	err := usersCollection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&existingUser)
	if err == nil {
		respondWithError(w, http.StatusConflict, "Email already registered")
		return
	} else if err != mongo.ErrNoDocuments {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	// Create user with email verified (no OTP required)
	now := time.Now()
	user := models.User{
		ID:            primitive.NewObjectID(),
		Email:         req.Email,
		PasswordHash:  hashedPassword,
		Role:          req.Role,
		EmailVerified: true, // Auto-verify email
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	_, err = usersCollection.InsertOne(ctx, user)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Generate JWT token immediately
	token, err := auth.GenerateJWT(user.ID.Hex(), user.Role)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "User registered successfully",
		"token":   token,
		"user": map[string]interface{}{
			"id":    user.ID.Hex(),
			"email": user.Email,
			"role":  user.Role,
		},
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	usersCollection := h.DB.Collection("users")
	var user models.User
	err := usersCollection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&user)

	if err == mongo.ErrNoDocuments {
		respondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	if !auth.CheckPasswordHash(req.Password, user.PasswordHash) {
		respondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Generate JWT token directly (no OTP verification required)
	token, err := auth.GenerateJWT(user.ID.Hex(), user.Role)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"id":    user.ID.Hex(),
			"email": user.Email,
			"role":  user.Role,
		},
	})
}

func (h *Handler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	var req models.OTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	otpsCollection := h.DB.Collection("otps")
	var otp models.OTP
	err := otpsCollection.FindOne(
		ctx,
		bson.M{"email": req.Email, "code": req.Code},
		options.FindOne().SetSort(bson.M{"created_at": -1}),
	).Decode(&otp)

	if err == mongo.ErrNoDocuments {
		respondWithError(w, http.StatusBadRequest, "Invalid OTP")
		return
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	if otp.Used {
		respondWithError(w, http.StatusBadRequest, "OTP already used")
		return
	}

	if time.Now().After(otp.ExpiresAt) {
		respondWithError(w, http.StatusBadRequest, "OTP expired")
		return
	}

	// Mark OTP as used
	_, err = otpsCollection.UpdateOne(
		ctx,
		bson.M{"_id": otp.ID},
		bson.M{"$set": bson.M{"used": true}},
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update OTP")
		return
	}

	// Verify user email
	usersCollection := h.DB.Collection("users")
	_, err = usersCollection.UpdateOne(
		ctx,
		bson.M{"email": req.Email},
		bson.M{"$set": bson.M{"email_verified": true, "updated_at": time.Now()}},
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to verify email")
		return
	}

	// Get user and generate JWT
	var user models.User
	err = usersCollection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	token, err := auth.GenerateJWT(user.ID.Hex(), user.Role)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Email verified successfully",
		"token":   token,
		"user": map[string]interface{}{
			"id":    user.ID.Hex(),
			"email": user.Email,
			"role":  user.Role,
		},
	})
}

func (h *Handler) ResendOTP(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if user exists
	usersCollection := h.DB.Collection("users")
	var user models.User
	err := usersCollection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		respondWithError(w, http.StatusNotFound, "User not found")
		return
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	// Generate new OTP
	otpCode := auth.GenerateOTP()
	expiresAt := auth.GetOTPExpiry()
	otp := models.OTP{
		ID:        primitive.NewObjectID(),
		Email:     req.Email,
		Code:      otpCode,
		ExpiresAt: expiresAt,
		Used:      false,
		CreatedAt: time.Now(),
	}

	otpsCollection := h.DB.Collection("otps")
	_, err = otpsCollection.InsertOne(ctx, otp)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to generate OTP")
		return
	}

	// In production, send OTP via email
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "OTP sent successfully",
		"otp":     otpCode, // Remove in production
	})
}
