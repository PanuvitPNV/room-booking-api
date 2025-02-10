package handlers

import (
	"github.com/panuvitpnv/room-booking-api/internal/dto/response"
	"github.com/panuvitpnv/room-booking-api/internal/models"
)

// Room related converters
func convertToRoomResponse(room models.Room) response.RoomResponse {
	return response.RoomResponse{
		RoomNum:  room.RoomNum,
		TypeID:   room.TypeID,
		RoomType: convertToRoomTypeResponse(room.RoomType),
	}
}

func convertToRoomTypeResponse(roomType models.RoomType) response.RoomTypeResponse {
	return response.RoomTypeResponse{
		TypeID:        roomType.TypeID,
		Name:          roomType.Name,
		Description:   roomType.Description,
		Area:          roomType.Area,
		Highlight:     roomType.Highlight,
		Facility:      roomType.Facility,
		PricePerNight: roomType.PricePerNight,
		Capacity:      roomType.Capacity,
	}
}

// Guest related converters
func convertToGuestResponse(guest models.Guest) response.GuestResponse {
	return response.GuestResponse{
		GuestID:     guest.GuestID,
		FirstName:   guest.FirstName,
		LastName:    guest.LastName,
		DateOfBirth: guest.DateOfBirth,
		Email:       guest.Email,
		Phone:       guest.Phone,
	}
}

// Booking related converters
func convertToBookingResponse(booking models.Booking) response.BookingResponse {
	return response.BookingResponse{
		BookingID:    booking.BookingID,
		RoomNum:      booking.RoomNum,
		GuestID:      booking.GuestID,
		CheckInDate:  booking.CheckInDate,
		CheckOutDate: booking.CheckOutDate,
		TotalPrice:   booking.TotalPrice,
		Room:         convertToRoomResponse(booking.Room),
		Guest:        convertToGuestResponse(booking.Guest),
	}
}

// Status related converters
func getStatusString(status *models.RoomStatus) string {
	if status == nil {
		return "Available"
	}
	return status.Status
}

func getAvailabilityStatus(available bool) string {
	if available {
		return "Available"
	}
	return "Not Available"
}
