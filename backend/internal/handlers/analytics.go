package handlers

import (
	"net/http"

	"seap/internal/middleware"
)

func (h *Handler) GetUserAnalytics(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

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

	h.DB.QueryRow(
		"SELECT COUNT(*) FROM campaigns WHERE user_id = $1",
		userID,
	).Scan(&stats.TotalCampaigns)

	h.DB.QueryRow(
		"SELECT COUNT(*) FROM campaigns WHERE user_id = $1 AND status = 'approved'",
		userID,
	).Scan(&stats.ApprovedCampaigns)

	h.DB.QueryRow(
		"SELECT COUNT(*) FROM campaigns WHERE user_id = $1 AND status = 'pending'",
		userID,
	).Scan(&stats.PendingCampaigns)

	h.DB.QueryRow(
		"SELECT COUNT(*) FROM campaigns WHERE user_id = $1 AND status = 'rejected'",
		userID,
	).Scan(&stats.RejectedCampaigns)

	h.DB.QueryRow(`
		SELECT COUNT(*) FROM events e
		JOIN campaigns c ON e.campaign_id = c.id
		WHERE c.user_id = $1 AND (e.event_type = 'link_opened' OR e.event_type = 'clicked')
	`, userID).Scan(&stats.TotalClicks)

	h.DB.QueryRow(`
		SELECT COUNT(*) FROM events e
		JOIN campaigns c ON e.campaign_id = c.id
		WHERE c.user_id = $1 AND e.event_type = 'form_submitted'
	`, userID).Scan(&stats.TotalSubmissions)

	h.DB.QueryRow(`
		SELECT COUNT(*) FROM events e
		JOIN campaigns c ON e.campaign_id = c.id
		WHERE c.user_id = $1 AND e.event_type = 'awareness_viewed'
	`, userID).Scan(&stats.TotalAwarenessViews)

	if stats.TotalClicks > 0 {
		stats.ConversionRate = float64(stats.TotalAwarenessViews) / float64(stats.TotalClicks) * 100
	}

	rows, err := h.DB.Query(`
		SELECT 
			c.id,
			c.title,
			c.status,
			COUNT(DISTINCT CASE WHEN e.event_type = 'link_opened' OR e.event_type = 'clicked' THEN e.id END) as clicks,
			COUNT(DISTINCT CASE WHEN e.event_type = 'form_submitted' THEN e.id END) as submissions,
			COUNT(DISTINCT CASE WHEN e.event_type = 'awareness_viewed' THEN e.id END) as awareness_views
		FROM campaigns c
		LEFT JOIN events e ON c.id = e.campaign_id
		WHERE c.user_id = $1
		GROUP BY c.id, c.title, c.status
		ORDER BY c.created_at DESC
	`, userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer rows.Close()

	type CampaignPerformance struct {
		ID             int    `json:"id"`
		Title          string `json:"title"`
		Status         string `json:"status"`
		Clicks         int    `json:"clicks"`
		Submissions    int    `json:"submissions"`
		AwarenessViews int    `json:"awareness_views"`
	}

	var campaigns []CampaignPerformance
	for rows.Next() {
		var cp CampaignPerformance
		err := rows.Scan(&cp.ID, &cp.Title, &cp.Status, &cp.Clicks, &cp.Submissions, &cp.AwarenessViews)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to scan campaign performance")
			return
		}
		campaigns = append(campaigns, cp)
	}

	rows, err = h.DB.Query(`
		SELECT DATE(e.created_at) as date, COUNT(*) as count
		FROM events e
		JOIN campaigns c ON e.campaign_id = c.id
		WHERE c.user_id = $1 AND e.created_at >= NOW() - INTERVAL '30 days'
		GROUP BY DATE(e.created_at)
		ORDER BY date DESC
	`, userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer rows.Close()

	type TimelineEntry struct {
		Date  string `json:"date"`
		Count int    `json:"count"`
	}

	var timeline []TimelineEntry
	for rows.Next() {
		var te TimelineEntry
		err := rows.Scan(&te.Date, &te.Count)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to scan timeline")
			return
		}
		timeline = append(timeline, te)
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"stats":     stats,
		"campaigns": campaigns,
		"timeline":  timeline,
	})
}

func (h *Handler) GetAdminAnalytics(w http.ResponseWriter, r *http.Request) {
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

	h.DB.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'user'").Scan(&stats.TotalUsers)
	h.DB.QueryRow("SELECT COUNT(*) FROM campaigns").Scan(&stats.TotalCampaigns)
	h.DB.QueryRow("SELECT COUNT(*) FROM campaigns WHERE status = 'approved'").Scan(&stats.ApprovedCampaigns)
	h.DB.QueryRow("SELECT COUNT(*) FROM campaigns WHERE status = 'pending'").Scan(&stats.PendingCampaigns)
	h.DB.QueryRow("SELECT COUNT(*) FROM campaigns WHERE status = 'rejected'").Scan(&stats.RejectedCampaigns)
	h.DB.QueryRow("SELECT COUNT(*) FROM events").Scan(&stats.TotalEvents)
	h.DB.QueryRow("SELECT COUNT(*) FROM events WHERE event_type = 'link_opened' OR event_type = 'clicked'").Scan(&stats.TotalClicks)
	h.DB.QueryRow("SELECT COUNT(*) FROM events WHERE event_type = 'awareness_viewed'").Scan(&stats.TotalConversions)

	if stats.TotalClicks > 0 {
		stats.AverageConversionRate = float64(stats.TotalConversions) / float64(stats.TotalClicks) * 100
	}

	rows, err := h.DB.Query(`
		SELECT status, COUNT(*) as count
		FROM campaigns
		GROUP BY status
	`)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer rows.Close()

	type StatusDistribution struct {
		Status string `json:"status"`
		Count  int    `json:"count"`
	}

	var distribution []StatusDistribution
	for rows.Next() {
		var sd StatusDistribution
		err := rows.Scan(&sd.Status, &sd.Count)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to scan distribution")
			return
		}
		distribution = append(distribution, sd)
	}

	rows, err = h.DB.Query(`
		SELECT DATE(created_at) as date, COUNT(*) as count
		FROM events
		WHERE created_at >= NOW() - INTERVAL '30 days'
		GROUP BY DATE(created_at)
		ORDER BY date DESC
	`)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer rows.Close()

	type TimelineEntry struct {
		Date  string `json:"date"`
		Count int    `json:"count"`
	}

	var timeline []TimelineEntry
	for rows.Next() {
		var te TimelineEntry
		err := rows.Scan(&te.Date, &te.Count)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to scan timeline")
			return
		}
		timeline = append(timeline, te)
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"stats":        stats,
		"distribution": distribution,
		"timeline":     timeline,
	})
}
