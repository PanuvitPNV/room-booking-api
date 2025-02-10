package request

import (
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

// ParseGetBookingsRequest extracts query parameters from the request and maps them to the struct.
func ParseGetBookingsRequest(c echo.Context) (*GetBookingsRequest, error) {
	var req GetBookingsRequest

	// Parse optional integer values
	if guestID := c.QueryParam("guest_id"); guestID != "" {
		id, err := strconv.Atoi(guestID)
		if err == nil {
			req.GuestID = &id
		}
	}
	if roomNum := c.QueryParam("room_num"); roomNum != "" {
		num, err := strconv.Atoi(roomNum)
		if err == nil {
			req.RoomNum = &num
		}
	}

	// Parse optional date values
	if fromDate := c.QueryParam("from_date"); fromDate != "" {
		parsedDate, err := time.Parse("2006-01-02", fromDate)
		if err == nil {
			t := parsedDate
			req.FromDate = &t
		}
	}
	if toDate := c.QueryParam("to_date"); toDate != "" {
		parsedDate, err := time.Parse("2006-01-02", toDate)
		if err == nil {
			t := parsedDate
			req.ToDate = &t
		}
	}

	// Parse required pagination values with defaults
	req.Page, _ = strconv.Atoi(c.QueryParam("page"))
	if req.Page == 0 {
		req.Page = 1 // Default to page 1
	}
	req.PageSize, _ = strconv.Atoi(c.QueryParam("page_size"))
	if req.PageSize == 0 {
		req.PageSize = 10 // Default to 10
	}

	return &req, nil
}
