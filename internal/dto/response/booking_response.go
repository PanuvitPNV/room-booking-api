package response

import (
	"time"
)

type BookingResponse struct {
	BookingID    int           `json:"booking_id"`
	RoomNum      int           `json:"room_num"`
	GuestID      int           `json:"guest_id"`
	CheckInDate  time.Time     `json:"check_in_date"`
	CheckOutDate time.Time     `json:"check_out_date"`
	TotalPrice   int           `json:"total_price"`
	Room         RoomResponse  `json:"room"`
	Guest        GuestResponse `json:"guest"`
}

type BookingDetailResponse struct {
	BookingID    int           `json:"booking_id"`
	RoomNum      int           `json:"room_num"`
	GuestID      int           `json:"guest_id"`
	CheckInDate  time.Time     `json:"check_in_date"`
	CheckOutDate time.Time     `json:"check_out_date"`
	TotalPrice   int           `json:"total_price"`
	Room         RoomResponse  `json:"room"`
	Guest        GuestResponse `json:"guest"`
	Status       string        `json:"status"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

type BookingAvailabilityResponse struct {
	RoomNum       int    `json:"room_num"`
	RoomType      string `json:"room_type"`
	Status        string `json:"status"`
	PricePerNight int    `json:"price_per_night"`
}

type BookingListResponse struct {
	Bookings []BookingResponse `json:"bookings"`
	Total    int               `json:"total"`
	Page     int               `json:"page"`
	LastPage int               `json:"last_page"`
}
