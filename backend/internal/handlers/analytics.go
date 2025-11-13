package handlers

import (
	"context"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"seap/internal/middleware"
)

func (h *Handler) GetUserAnalytics(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value(middleware.UserIDKey).(string)
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	campaignsCollection := h.DB.Collection("campaigns")
	eventsCollection := h.DB.Collection("events")

	var stats struct {
		TotalCampaigns      int     `json:"total_campaigns"`
		ApprovedCampaigns   int     `json:"approved_campaigns"`
		PendingCampaigns    int     `json:"pending_campaigns"`
		RejectedCampaigns   int     `json:"rejected_campaigns"`
		TotalClicks         int     `json:"total_clicks"`
		TotalSubmissions    int     `json:"total_submissions"`
		TotalAwarenessViews int     `json:"total_awareness_views"`
		ConversionRate      float64 `json:"conversion_rate"`
	}

	// Count campaigns
	totalCampaigns, _ := campaignsCollection.CountDocuments(ctx, bson.M{"user_id": userID})
	stats.TotalCampaigns = int(totalCampaigns)
	approvedCampaigns, _ := campaignsCollection.CountDocuments(ctx, bson.M{"user_id": userID, "status": "approved"})
	stats.ApprovedCampaigns = int(approvedCampaigns)
	pendingCampaigns, _ := campaignsCollection.CountDocuments(ctx, bson.M{"user_id": userID, "status": "pending"})
	stats.PendingCampaigns = int(pendingCampaigns)
	rejectedCampaigns, _ := campaignsCollection.CountDocuments(ctx, bson.M{"user_id": userID, "status": "rejected"})
	stats.RejectedCampaigns = int(rejectedCampaigns)

	// Get campaign IDs
	campaignCursor, _ := campaignsCollection.Find(ctx, bson.M{"user_id": userID})
	var campaignIDs []primitive.ObjectID
	for campaignCursor.Next(ctx) {
		var campaign struct {
			ID primitive.ObjectID `bson:"_id"`
		}
		if err := campaignCursor.Decode(&campaign); err == nil {
			campaignIDs = append(campaignIDs, campaign.ID)
		}
	}
	campaignCursor.Close(ctx)

	if len(campaignIDs) > 0 {
		// Count clicks (only link_opened events - clicked is not used)
		totalClicks, _ := eventsCollection.CountDocuments(ctx, bson.M{
			"campaign_id": bson.M{"$in": campaignIDs},
			"event_type":  "link_opened",
		})
		stats.TotalClicks = int(totalClicks)

		// Count submissions
		totalSubmissions, _ := eventsCollection.CountDocuments(ctx, bson.M{
			"campaign_id": bson.M{"$in": campaignIDs},
			"event_type":  "form_submitted",
		})
		stats.TotalSubmissions = int(totalSubmissions)

		// Count awareness views
		totalAwarenessViews, _ := eventsCollection.CountDocuments(ctx, bson.M{
			"campaign_id": bson.M{"$in": campaignIDs},
			"event_type":  "awareness_viewed",
		})
		stats.TotalAwarenessViews = int(totalAwarenessViews)
	}

	if stats.TotalClicks > 0 {
		stats.ConversionRate = float64(stats.TotalAwarenessViews) / float64(stats.TotalClicks) * 100
	}

	// Get campaign performance
	type CampaignPerformance struct {
		ID             string `json:"id"`
		Title          string `json:"title"`
		Status         string `json:"status"`
		Clicks         int    `json:"clicks"`
		Submissions    int    `json:"submissions"`
		AwarenessViews int    `json:"awareness_views"`
	}

	var campaigns []CampaignPerformance
	campaignCursor, _ = campaignsCollection.Find(ctx, bson.M{"user_id": userID})
	for campaignCursor.Next(ctx) {
		var campaign struct {
			ID     primitive.ObjectID `bson:"_id"`
			Title  string             `bson:"title"`
			Status string             `bson:"status"`
		}
		if err := campaignCursor.Decode(&campaign); err != nil {
			continue
		}

		clicks, _ := eventsCollection.CountDocuments(ctx, bson.M{
			"campaign_id": campaign.ID,
			"event_type":  "link_opened",
		})
		submissions, _ := eventsCollection.CountDocuments(ctx, bson.M{
			"campaign_id": campaign.ID,
			"event_type":  "form_submitted",
		})
		awarenessViews, _ := eventsCollection.CountDocuments(ctx, bson.M{
			"campaign_id": campaign.ID,
			"event_type":  "awareness_viewed",
		})

		campaigns = append(campaigns, CampaignPerformance{
			ID:             campaign.ID.Hex(),
			Title:          campaign.Title,
			Status:         campaign.Status,
			Clicks:         int(clicks),
			Submissions:    int(submissions),
			AwarenessViews: int(awarenessViews),
		})
	}
	campaignCursor.Close(ctx)

	// Get timeline (last 30 days)
	type TimelineEntry struct {
		Date  string `json:"date"`
		Count int    `json:"count"`
	}

	var timeline []TimelineEntry
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	if len(campaignIDs) > 0 {
		pipeline := []bson.M{
			{
				"$match": bson.M{
					"campaign_id": bson.M{"$in": campaignIDs},
					"created_at":  bson.M{"$gte": thirtyDaysAgo},
				},
			},
			{
				"$group": bson.M{
					"_id": bson.M{
						"$dateToString": bson.M{
							"format": "%Y-%m-%d",
							"date":   "$created_at",
						},
					},
					"count": bson.M{"$sum": 1},
				},
			},
			{
				"$sort": bson.M{"_id": -1},
			},
		}

		cursor, err := eventsCollection.Aggregate(ctx, pipeline)
		if err == nil {
			for cursor.Next(ctx) {
				var result struct {
					ID    string `bson:"_id"`
					Count int    `bson:"count"`
				}
				if err := cursor.Decode(&result); err == nil {
					timeline = append(timeline, TimelineEntry{
						Date:  result.ID,
						Count: result.Count,
					})
				}
			}
			cursor.Close(ctx)
		}
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"stats":     stats,
		"campaigns": campaigns,
		"timeline":  timeline,
	})
}

func (h *Handler) GetAdminAnalytics(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	usersCollection := h.DB.Collection("users")
	campaignsCollection := h.DB.Collection("campaigns")
	eventsCollection := h.DB.Collection("events")

	var stats struct {
		TotalUsers            int     `json:"total_users"`
		TotalCampaigns        int     `json:"total_campaigns"`
		ApprovedCampaigns     int     `json:"approved_campaigns"`
		PendingCampaigns      int     `json:"pending_campaigns"`
		RejectedCampaigns     int     `json:"rejected_campaigns"`
		TotalEvents           int     `json:"total_events"`
		TotalClicks           int     `json:"total_clicks"`
		TotalConversions      int     `json:"total_conversions"`
		AverageConversionRate float64 `json:"average_conversion_rate"`
	}

	totalUsers, _ := usersCollection.CountDocuments(ctx, bson.M{"role": "user"})
	stats.TotalUsers = int(totalUsers)
	totalCampaigns, _ := campaignsCollection.CountDocuments(ctx, bson.M{})
	stats.TotalCampaigns = int(totalCampaigns)
	approvedCampaigns, _ := campaignsCollection.CountDocuments(ctx, bson.M{"status": "approved"})
	stats.ApprovedCampaigns = int(approvedCampaigns)
	pendingCampaigns, _ := campaignsCollection.CountDocuments(ctx, bson.M{"status": "pending"})
	stats.PendingCampaigns = int(pendingCampaigns)
	rejectedCampaigns, _ := campaignsCollection.CountDocuments(ctx, bson.M{"status": "rejected"})
	stats.RejectedCampaigns = int(rejectedCampaigns)
	totalEvents, _ := eventsCollection.CountDocuments(ctx, bson.M{})
	stats.TotalEvents = int(totalEvents)
	totalClicks, _ := eventsCollection.CountDocuments(ctx, bson.M{
		"event_type": "link_opened",
	})
	stats.TotalClicks = int(totalClicks)
	totalConversions, _ := eventsCollection.CountDocuments(ctx, bson.M{
		"event_type": "awareness_viewed",
	})
	stats.TotalConversions = int(totalConversions)

	if stats.TotalClicks > 0 {
		stats.AverageConversionRate = float64(stats.TotalConversions) / float64(stats.TotalClicks) * 100
	}

	// Status distribution
	type StatusDistribution struct {
		Status string `json:"status"`
		Count  int    `json:"count"`
	}

	var distribution []StatusDistribution
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id":   "$status",
				"count": bson.M{"$sum": 1},
			},
		},
	}

	cursor, err := campaignsCollection.Aggregate(ctx, pipeline)
	if err == nil {
		for cursor.Next(ctx) {
			var result struct {
				ID    string `bson:"_id"`
				Count int    `bson:"count"`
			}
			if err := cursor.Decode(&result); err == nil {
				distribution = append(distribution, StatusDistribution{
					Status: result.ID,
					Count:  result.Count,
				})
			}
		}
		cursor.Close(ctx)
	}

	// Timeline (last 30 days)
	type TimelineEntry struct {
		Date  string `json:"date"`
		Count int    `json:"count"`
	}

	var timeline []TimelineEntry
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	pipeline = []bson.M{
		{
			"$match": bson.M{
				"created_at": bson.M{"$gte": thirtyDaysAgo},
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"$dateToString": bson.M{
						"format": "%Y-%m-%d",
						"date":   "$created_at",
					},
				},
				"count": bson.M{"$sum": 1},
			},
		},
		{
			"$sort": bson.M{"_id": -1},
		},
	}

	cursor, err = eventsCollection.Aggregate(ctx, pipeline)
	if err == nil {
		for cursor.Next(ctx) {
			var result struct {
				ID    string `bson:"_id"`
				Count int    `bson:"count"`
			}
			if err := cursor.Decode(&result); err == nil {
				timeline = append(timeline, TimelineEntry{
					Date:  result.ID,
					Count: result.Count,
				})
			}
		}
		cursor.Close(ctx)
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"stats":        stats,
		"distribution": distribution,
		"timeline":     timeline,
	})
}
