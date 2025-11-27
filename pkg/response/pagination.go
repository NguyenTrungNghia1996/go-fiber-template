package response

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// ListData is a conventional data shape for list endpoints.
// It nests pagination metadata inside the data field, keeping the top-level
// envelope consistent as: data + message + status.
type ListData[T any] struct {
	Items []T   `json:"items"`
	Page  int64 `json:"page"`
	Limit int64 `json:"limit"`
	Total int64 `json:"total"`
}

// ParsePageLimit extracts page and limit from query parameters.
// Behavior:
// - If page is missing or equals 0 -> return (0, 0) meaning "return all".
// - Otherwise defaults: page=1, limit=10 with limit clamped to [1,100].
func ParsePageLimit(c *fiber.Ctx) (int64, int64) {
	// Detect page all-data mode
	if v := c.Query("page"); v == "" || v == "0" {
		return 0, 0
	}
	// defaults
	var (
		page  int64 = 1
		limit int64 = 10
	)
	if v := c.Query("page"); v != "" {
		if p, err := strconv.ParseInt(v, 10, 64); err == nil && p > 0 {
			page = p
		}
	}
	if v := c.Query("limit"); v != "" {
		if l, err := strconv.ParseInt(v, 10, 64); err == nil && l > 0 {
			limit = l
		}
	}
	if limit > 100 {
		limit = 100
	}
	return page, limit
}
