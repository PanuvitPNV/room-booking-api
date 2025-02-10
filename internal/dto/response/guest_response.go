package response

import "time"

type GuestResponse struct {
	GuestID     int       `json:"guest_id"`
	FirstName   string    `json:"f_name"`
	LastName    string    `json:"l_name"`
	DateOfBirth time.Time `json:"date_of_birth"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone"`
}

type GuestDetailResponse struct {
	GuestID     int       `json:"guest_id"`
	FirstName   string    `json:"f_name"`
	LastName    string    `json:"l_name"`
	DateOfBirth time.Time `json:"date_of_birth"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type GuestBookingHistoryResponse struct {
	GuestID  int               `json:"guest_id"`
	Bookings []BookingResponse `json:"bookings"`
	Total    int               `json:"total"`
	Page     int               `json:"page"`
	LastPage int               `json:"last_page"`
}

type GuestListResponse struct {
	Guests   []GuestResponse `json:"guests"`
	Total    int             `json:"total"`
	Page     int             `json:"page"`
	LastPage int             `json:"last_page"`
}
