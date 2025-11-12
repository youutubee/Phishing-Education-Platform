package models

import "time"

type User struct {
	ID            int       `json:"id"`
	Email         string    `json:"email"`
	PasswordHash  string    `json:"-"`
	Role          string    `json:"role"`
	EmailVerified bool      `json:"email_verified"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Campaign struct {
	ID            int       `json:"id"`
	UserID        int       `json:"user_id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	EmailText     string    `json:"email_text"`
	LandingPageURL string   `json:"landing_page_url"`
	TrackingToken string    `json:"tracking_token"`
	Status        string    `json:"status"` // pending, approved, rejected
	ExpiryDate    *time.Time `json:"expiry_date"`
	AdminComment  string     `json:"admin_comment"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type Event struct {
	ID         int       `json:"id"`
	CampaignID int       `json:"campaign_id"`
	EventType  string    `json:"event_type"` // link_opened, clicked, form_submitted, awareness_viewed
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	CreatedAt  time.Time `json:"created_at"`
}

type AuditLog struct {
	ID          int       `json:"id"`
	AdminID     int       `json:"admin_id"`
	Action      string    `json:"action"`
	ResourceType string   `json:"resource_type"`
	ResourceID  *int      `json:"resource_id"`
	Details     string    `json:"details"`
	CreatedAt   time.Time `json:"created_at"`
}

type OTP struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `json:"used"`
	CreatedAt time.Time `json:"created_at"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type OTPRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type CampaignRequest struct {
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	EmailText     string     `json:"email_text"`
	LandingPageURL string    `json:"landing_page_url"`
	ExpiryDate    *time.Time `json:"expiry_date"`
}

type CampaignApprovalRequest struct {
	Comment string `json:"comment"`
}

type LeaderboardEntry struct {
	UserID           int    `json:"user_id"`
	Email            string `json:"email"`
	TotalClicks      int    `json:"total_clicks"`
	TotalConversions int    `json:"total_conversions"`
	TotalCampaigns   int    `json:"total_campaigns"`
	RejectedCount    int    `json:"rejected_count"`
	Score            int    `json:"score"`
}

