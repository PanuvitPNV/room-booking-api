package request

import "time"

// CreateRoomRequest aligns with RoomRepository.Create
type CreateRoomRequest struct {
	RoomNum int `json:"room_num" validate:"required"`
	TypeID  int `json:"type_id" validate:"required"`
}

// UpdateRoomRequest aligns with RoomRepository.Update
type UpdateRoomRequest struct {
	RoomNum int `json:"room_num" validate:"required"`
	TypeID  int `json:"type_id" validate:"required"`
}

// GetAvailableRoomsRequest aligns with RoomRepository.GetAvailableRooms
type GetAvailableRoomsRequest struct {
	TypeID       *int      `json:"type_id,omitempty"`
	CheckInDate  time.Time `json:"check_in_date" validate:"required"`
	CheckOutDate time.Time `json:"check_out_date" validate:"required"`
}

// GetRoomStatusRequest aligns with RoomStatusRepository.GetRoomStatusRange
type GetRoomStatusRequest struct {
	RoomNum   int       `json:"room_num" validate:"required"`
	StartDate time.Time `json:"start_date" validate:"required"`
	EndDate   time.Time `json:"end_date" validate:"required"`
}

// CreateRoomTypeRequest aligns with RoomTypeRepository.Create
type CreateRoomTypeRequest struct {
	Name          string `json:"name" validate:"required"`
	Description   string `json:"description"`
	Area          int    `json:"area" validate:"required,gt=0"`
	Highlight     string `json:"highlight"`
	Facility      string `json:"facility"`
	PricePerNight int    `json:"price_per_night" validate:"required,gt=0"`
	Capacity      int    `json:"capacity" validate:"required,gt=0"`
}

// UpdateRoomTypeRequest aligns with RoomTypeRepository.Update
type UpdateRoomTypeRequest struct {
	TypeID        int    `json:"type_id" validate:"required"`
	Name          string `json:"name,omitempty"`
	Description   string `json:"description,omitempty"`
	Area          int    `json:"area,omitempty"`
	Highlight     string `json:"highlight,omitempty"`
	Facility      string `json:"facility,omitempty"`
	PricePerNight int    `json:"price_per_night,omitempty"`
	Capacity      int    `json:"capacity,omitempty"`
}
