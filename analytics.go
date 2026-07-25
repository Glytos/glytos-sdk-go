package glytos

import (
	"context"
	"net/url"
	"strconv"
)

// AnalyticsService exposes usage and activity analytics.
type AnalyticsService struct{ client *Client }

// Overview returns a high-level usage and cost overview for the last days days
// (1-90; the server defaults to 14). Pass days <= 0 to use the server default.
func (s *AnalyticsService) Overview(ctx context.Context, days int) (*AnalyticsOverview, error) {
	query := url.Values{}
	if days > 0 {
		query.Set("days", strconv.Itoa(days))
	}
	var out AnalyticsOverview
	err := s.client.do(ctx, "GET", "/analytics/overview", nil, query, &out)
	return &out, err
}
