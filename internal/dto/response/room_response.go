package response

import "time"

type RoomTypeResponse struct {
	TypeID        int    `json:"type_id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Area          int    `json:"area"`
	Highlight     string `json:"highlight"`
	Facility      string `json:"facility"`
	PricePerNight int    `json:"price_per_night"`
	Capacity      int    `json:"capacity"`
}

type RoomResponse struct {
	RoomNum  int              `json:"room_num"`
	TypeID   int              `json:"type_id"`
	RoomType RoomTypeResponse `json:"room_type"`
}

type RoomStatusInfo struct {
	Status    string     `json:"status"`
	BookingID *int       `json:"booking_id,omitempty"`
	CheckIn   *time.Time `json:"check_in,omitempty"`
	CheckOut  *time.Time `json:"check_out,omitempty"`
}

type RoomDetailResponse struct {
	RoomNum       int              `json:"room_num"`
	TypeID        int              `json:"type_id"`
	RoomType      RoomTypeResponse `json:"room_type"`
	CurrentStatus RoomStatusInfo   `json:"current_status"`
}

type RoomStatusResponse struct {
	RoomNum   int       `json:"room_num"`
	Calendar  time.Time `json:"calendar"`
	Status    string    `json:"status"`
	BookingID *int      `json:"booking_id,omitempty"`
}

type RoomListResponse struct {
	Rooms    []RoomResponse `json:"rooms"`
	Total    int            `json:"total"`
	Page     int            `json:"page"`
	LastPage int            `json:"last_page"`
}

type RoomTypeListResponse struct {
	RoomTypes []RoomTypeResponse `json:"room_types"`
	Total     int                `json:"total"`
	Page      int                `json:"page"`
	LastPage  int                `json:"last_page"`
}

type RoomAvailabilityResponse struct {
	RoomNum       int    `json:"room_num"`
	RoomType      string `json:"room_type"`
	Status        string `json:"status"`
	PricePerNight int    `json:"price_per_night"`
}
