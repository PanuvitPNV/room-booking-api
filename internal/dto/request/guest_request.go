package request

import "time"

// CreateGuestRequest aligns with GuestRepository.Create
type CreateGuestRequest struct {
	FirstName   string    `json:"f_name" validate:"required"`
	LastName    string    `json:"l_name" validate:"required"`
	DateOfBirth time.Time `json:"date_of_birth" validate:"required"`
	Email       string    `json:"email" validate:"required,email"`
	Phone       string    `json:"phone" validate:"required"`
}

// UpdateGuestRequest aligns with GuestRepository.Update
type UpdateGuestRequest struct {
	GuestID     int       `json:"guest_id" validate:"required"`
	FirstName   string    `json:"f_name,omitempty"`
	LastName    string    `json:"l_name,omitempty"`
	DateOfBirth time.Time `json:"date_of_birth,omitempty"`
	Email       string    `json:"email,omitempty" validate:"omitempty,email"`
	Phone       string    `json:"phone,omitempty"`
}

// GetGuestRequest for fetching guest details with optional booking history
type GetGuestRequest struct {
	GuestID         int  `json:"guest_id" validate:"required"`
	IncludeBookings bool `json:"include_bookings"`
}

// ListGuestsRequest for paginated guest listing
type ListGuestsRequest struct {
	Page     int    `json:"page" validate:"required,min=1" default:"1"`
	PageSize int    `json:"page_size" validate:"required,min=1,max=100" default:"10"`
	Search   string `json:"search,omitempty"`
}
