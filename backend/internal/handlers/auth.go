package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

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

	// Check if user exists
	var existingID int
	err := h.DB.QueryRow("SELECT id FROM users WHERE email = $1", req.Email).Scan(&existingID)
	if err == nil {
		respondWithError(w, http.StatusConflict, "Email already registered")
		return
	} else if err != sql.ErrNoRows {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	// Create user
	var userID int
	err = h.DB.QueryRow(
		"INSERT INTO users (email, password_hash, role, email_verified) VALUES ($1, $2, $3, $4) RETURNING id",
		req.Email, hashedPassword, req.Role, false,
	).Scan(&userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Generate and store OTP
	otpCode := auth.GenerateOTP()
	expiresAt := auth.GetOTPExpiry()
	_, err = h.DB.Exec(
		"INSERT INTO otps (email, code, expires_at) VALUES ($1, $2, $3)",
		req.Email, otpCode, expiresAt,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to generate OTP")
		return
	}

	// In production, send OTP via email
	// For now, return it in response (remove in production)
	respondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "User registered successfully. Please verify your email with OTP.",
		"otp":     otpCode, // Remove this in production
		"user_id": userID,
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	var user models.User
	err := h.DB.QueryRow(
		"SELECT id, email, password_hash, role, email_verified FROM users WHERE email = $1",
		req.Email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.EmailVerified)

	if err == sql.ErrNoRows {
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

	if !user.EmailVerified {
		// Generate new OTP
		otpCode := auth.GenerateOTP()
		expiresAt := auth.GetOTPExpiry()
		_, err = h.DB.Exec(
			"INSERT INTO otps (email, code, expires_at) VALUES ($1, $2, $3)",
			req.Email, otpCode, expiresAt,
		)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to generate OTP")
			return
		}

		respondWithJSON(w, http.StatusOK, map[string]interface{}{
			"message": "Email not verified. Please verify with OTP.",
			"otp":     otpCode, // Remove in production
		})
		return
	}

	// Generate JWT
	token, err := auth.GenerateJWT(user.ID, user.Role)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"id":    user.ID,
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

	var otp models.OTP
	err := h.DB.QueryRow(
		"SELECT id, email, code, expires_at, used FROM otps WHERE email = $1 AND code = $2 ORDER BY created_at DESC LIMIT 1",
		req.Email, req.Code,
	).Scan(&otp.ID, &otp.Email, &otp.Code, &otp.ExpiresAt, &otp.Used)

	if err == sql.ErrNoRows {
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
	_, err = h.DB.Exec("UPDATE otps SET used = TRUE WHERE id = $1", otp.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update OTP")
		return
	}

	// Verify user email
	_, err = h.DB.Exec("UPDATE users SET email_verified = TRUE WHERE email = $1", req.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to verify email")
		return
	}

	// Get user and generate JWT
	var user models.User
	err = h.DB.QueryRow(
		"SELECT id, email, role FROM users WHERE email = $1",
		req.Email,
	).Scan(&user.ID, &user.Email, &user.Role)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	token, err := auth.GenerateJWT(user.ID, user.Role)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Email verified successfully",
		"token":   token,
		"user": map[string]interface{}{
			"id":    user.ID,
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

	// Check if user exists
	var userID int
	err := h.DB.QueryRow("SELECT id FROM users WHERE email = $1", req.Email).Scan(&userID)
	if err == sql.ErrNoRows {
		respondWithError(w, http.StatusNotFound, "User not found")
		return
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	// Generate new OTP
	otpCode := auth.GenerateOTP()
	expiresAt := auth.GetOTPExpiry()
	_, err = h.DB.Exec(
		"INSERT INTO otps (email, code, expires_at) VALUES ($1, $2, $3)",
		req.Email, otpCode, expiresAt,
	)
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

