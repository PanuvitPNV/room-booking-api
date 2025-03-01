package models

import "github.com/panuvitpnv/room-booking-api/internal/models"

// RoomFacilityResponse is a custom response struct for room facilities
type RoomFacilityResponse struct {
	TypeID     int              `json:"type_id"`
	FacilityID int              `json:"fac_id"`
	Facility   FacilityResponse `json:"facility"`
}

// FacilityResponse is a custom response struct for facilities
type FacilityResponse struct {
	FacilityID int    `json:"fac_id"`
	Name       string `json:"name"`
}

// RoomTypeResponse is a custom response struct for room types
type RoomTypeResponse struct {
	TypeID         int                    `json:"type_id"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	Area           int                    `json:"area"`
	PricePerNight  int                    `json:"price_per_night"`
	NoOfGuest      int                    `json:"noOfGuest"`
	RoomFacilities []RoomFacilityResponse `json:"facilities"`
}

// RoomResponse is a custom response struct for rooms
type RoomResponse struct {
	RoomNum  int              `json:"room_num"`
	TypeID   int              `json:"type_id"`
	RoomType RoomTypeResponse `json:"room_type"`
}

// Convert database models to response models

// ConvertToRoomResponse converts a Room model to a RoomResponse
func ConvertToRoomResponse(room models.Room) RoomResponse {
	return RoomResponse{
		RoomNum:  room.RoomNum,
		TypeID:   room.TypeID,
		RoomType: ConvertToRoomTypeResponse(room.RoomType),
	}
}

// ConvertToRoomTypeResponse converts a RoomType model to a RoomTypeResponse
func ConvertToRoomTypeResponse(roomType models.RoomType) RoomTypeResponse {
	facilities := make([]RoomFacilityResponse, 0)

	for _, rf := range roomType.RoomFacilities {
		facilities = append(facilities, RoomFacilityResponse{
			TypeID:     rf.TypeID,
			FacilityID: rf.FacilityID,
			Facility: FacilityResponse{
				FacilityID: rf.FacilityID,
				Name:       rf.Facility.Name,
			},
		})
	}

	return RoomTypeResponse{
		TypeID:         roomType.TypeID,
		Name:           roomType.Name,
		Description:    roomType.Description,
		Area:           roomType.Area,
		PricePerNight:  roomType.PricePerNight,
		NoOfGuest:      roomType.NoOfGuest,
		RoomFacilities: facilities,
	}
}

// ConvertRoomsToResponse converts a slice of Room models to a slice of RoomResponses
func ConvertRoomsToResponse(rooms []models.Room) []RoomResponse {
	response := make([]RoomResponse, len(rooms))

	for i, room := range rooms {
		response[i] = ConvertToRoomResponse(room)
	}

	return response
}

// ConvertRoomTypesToResponse converts a slice of RoomType models to a slice of RoomTypeResponses
func ConvertRoomTypesToResponse(roomTypes []models.RoomType) []RoomTypeResponse {
	response := make([]RoomTypeResponse, len(roomTypes))

	for i, roomType := range roomTypes {
		response[i] = ConvertToRoomTypeResponse(roomType)
	}

	return response
}
