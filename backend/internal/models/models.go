package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email         string             `bson:"email" json:"email"`
	PasswordHash  string             `bson:"password_hash" json:"-"`
	Role          string             `bson:"role" json:"role"`
	EmailVerified bool               `bson:"email_verified" json:"email_verified"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
}

type Campaign struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID         primitive.ObjectID `bson:"user_id" json:"user_id"`
	Title          string             `bson:"title" json:"title"`
	Description    string             `bson:"description" json:"description"`
	EmailText      string             `bson:"email_text" json:"email_text"`
	LandingPageURL string             `bson:"landing_page_url" json:"landing_page_url"`
	TrackingToken  string             `bson:"tracking_token" json:"tracking_token"`
	Status         string             `bson:"status" json:"status"` // pending, approved, rejected
	ExpiryDate     *time.Time         `bson:"expiry_date,omitempty" json:"expiry_date"`
	AdminComment   string             `bson:"admin_comment,omitempty" json:"admin_comment"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`
}

type Event struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CampaignID primitive.ObjectID `bson:"campaign_id" json:"campaign_id"`
	EventType  string             `bson:"event_type" json:"event_type"` // link_opened, clicked, form_submitted, awareness_viewed
	IPAddress  string             `bson:"ip_address,omitempty" json:"ip_address"`
	UserAgent  string             `bson:"user_agent,omitempty" json:"user_agent"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
}

type AuditLog struct {
	ID           primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	AdminID      primitive.ObjectID  `bson:"admin_id" json:"admin_id"`
	Action       string              `bson:"action" json:"action"`
	ResourceType string              `bson:"resource_type" json:"resource_type"`
	ResourceID   *primitive.ObjectID `bson:"resource_id,omitempty" json:"resource_id"`
	Details      string              `bson:"details" json:"details"`
	CreatedAt    time.Time           `bson:"created_at" json:"created_at"`
}

type OTP struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email     string             `bson:"email" json:"email"`
	Code      string             `bson:"code" json:"code"`
	ExpiresAt time.Time          `bson:"expires_at" json:"expires_at"`
	Used      bool               `bson:"used" json:"used"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
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
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	EmailText      string     `json:"email_text"`
	LandingPageURL string     `json:"landing_page_url"`
	ExpiryDate     *time.Time `json:"expiry_date"`
}

type CampaignApprovalRequest struct {
	Comment string `json:"comment"`
}

type LeaderboardEntry struct {
	UserID           primitive.ObjectID `json:"user_id"`
	Email            string             `json:"email"`
	TotalClicks      int                `json:"total_clicks"`
	TotalConversions int                `json:"total_conversions"`
	TotalCampaigns   int                `json:"total_campaigns"`
	RejectedCount    int                `json:"rejected_count"`
	Score            int                `json:"score"`
}
