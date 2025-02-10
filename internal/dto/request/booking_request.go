package request

import "time"

// CreateBookingRequest aligns with BookingRepository.Create
type CreateBookingRequest struct {
	RoomNum      int       `json:"room_num" validate:"required"`
	GuestID      int       `json:"guest_id" validate:"required"`
	CheckInDate  time.Time `json:"check_in_date" validate:"required"`
	CheckOutDate time.Time `json:"check_out_date" validate:"required"`
}

// UpdateBookingRequest aligns with BookingRepository.Update
type UpdateBookingRequest struct {
	BookingID    int       `json:"booking_id" validate:"required"`
	CheckInDate  time.Time `json:"check_in_date,omitempty"`
	CheckOutDate time.Time `json:"check_out_date,omitempty"`
}

// CheckAvailabilityRequest aligns with BookingRepository.CheckRoomAvailability
type CheckAvailabilityRequest struct {
	RoomNum      int       `json:"room_num" validate:"required"`
	CheckInDate  time.Time `json:"check_in_date" validate:"required"`
	CheckOutDate time.Time `json:"check_out_date" validate:"required"`
}

// GetBookingsRequest for listing with optional filters
type GetBookingsRequest struct {
	GuestID  *int       `json:"guest_id,omitempty"`
	RoomNum  *int       `json:"room_num,omitempty"`
	FromDate *time.Time `json:"from_date,omitempty"`
	ToDate   *time.Time `json:"to_date,omitempty"`
	Page     int        `json:"page" validate:"required,min=1" default:"1"`
	PageSize int        `json:"page_size" validate:"required,min=1,max=100" default:"10"`
}
