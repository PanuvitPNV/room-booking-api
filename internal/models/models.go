package models

import "time"

// RoomType represents different types of rooms available in the hotel
type RoomType struct {
	TypeID         int            `gorm:"primaryKey;column:type_id" json:"type_id"`
	Name           string         `gorm:"unique;not null" json:"name"`
	Description    string         `json:"description"`
	Area           int            `gorm:"not null" json:"area"`
	PricePerNight  int            `gorm:"not null;column:price_per_night" json:"price_per_night"`
	NoOfGuest      int            `gorm:"not null;column:no_of_guest" json:"noOfGuest"`
	Rooms          []Room         `gorm:"foreignKey:TypeID" json:"rooms,omitempty"`
	RoomFacilities []RoomFacility `gorm:"foreignKey:TypeID" json:"room_facilities,omitempty"`
}

// Facility represents amenities available in rooms
type Facility struct {
	FacilityID     int            `gorm:"primaryKey;column:fac_id" json:"fac_id"`
	Name           string         `gorm:"unique;not null" json:"name"`
	RoomFacilities []RoomFacility `gorm:"foreignKey:FacilityID" json:"room_facilities,omitempty"`
}

// RoomFacility maps facilities to room types
type RoomFacility struct {
	TypeID     int      `gorm:"primaryKey;column:type_id" json:"type_id"`
	FacilityID int      `gorm:"primaryKey;column:fac_id" json:"fac_id"`
	RoomType   RoomType `gorm:"foreignKey:TypeID;references:TypeID" json:"-"`
	Facility   Facility `gorm:"foreignKey:FacilityID;references:FacilityID" json:"facility"` // Include this in JSON
}

// Room represents an actual room in the hotel
type Room struct {
	RoomNum  int          `gorm:"primaryKey;column:room_num" json:"room_num"`
	TypeID   int          `gorm:"not null" json:"type_id"`
	RoomType RoomType     `gorm:"foreignKey:TypeID;references:TypeID" json:"room_type,omitempty"`
	Bookings []Booking    `gorm:"foreignKey:RoomNum" json:"bookings,omitempty"`
	Statuses []RoomStatus `gorm:"foreignKey:RoomNum" json:"statuses,omitempty"`
}

// Booking represents a reservation for a room
type Booking struct {
	BookingID    int          `gorm:"primaryKey;column:booking_id" json:"booking_id"`
	BookingName  string       `gorm:"column:booking_name;not null" json:"booking_name" validate:"required"`
	RoomNum      int          `gorm:"not null" json:"room_num" validate:"required"`
	CheckInDate  time.Time    `gorm:"not null;column:check_in_date" json:"check_in_date" validate:"required"`
	CheckOutDate time.Time    `gorm:"not null;column:check_out_date" json:"check_out_date" validate:"required"`
	BookingDate  time.Time    `gorm:"not null;column:booking_date;default:CURRENT_TIMESTAMP" json:"booking_date"`
	TotalPrice   int          `gorm:"not null;column:total_price" json:"total_price"`
	CreatedAt    time.Time    `gorm:"autoCreateTime" json:"-"`
	UpdatedAt    time.Time    `gorm:"autoUpdateTime" json:"-"`
	Room         Room         `gorm:"foreignKey:RoomNum;references:RoomNum" json:"room,omitempty"`
	Receipt      *Receipt     `gorm:"foreignKey:BookingID" json:"receipt,omitempty"`
	Statuses     []RoomStatus `gorm:"foreignKey:BookingID" json:"statuses,omitempty"`
}

// Receipt represents payment confirmation for a booking
type Receipt struct {
	ReceiptID     int       `gorm:"primaryKey;column:receipt_id" json:"receipt_id"`
	BookingID     int       `gorm:"not null;column:booking_id" json:"booking_id"`
	PaymentDate   time.Time `gorm:"not null;column:payment_date" json:"payment_date"`
	PaymentMethod string    `gorm:"column:payment_method;not null" json:"payment_method" validate:"required,oneof=Credit Debit Bank Transfer"`
	Amount        int       `gorm:"not null;column:amount" json:"amount" validate:"required"`
	IssueDate     time.Time `gorm:"not null;column:issue_date;default:CURRENT_TIMESTAMP" json:"issue_date"`
	Booking       Booking   `gorm:"foreignKey:BookingID;references:BookingID" json:"-"`
}

// RoomStatus represents the availability status of a room on a specific date
type RoomStatus struct {
	RoomNum   int       `gorm:"primaryKey;column:room_num" json:"room_num"`
	Calendar  time.Time `gorm:"primaryKey;column:calendar;type:date" json:"calendar"`
	Status    string    `gorm:"not null;column:status;default:Available" json:"status" validate:"required,oneof=Available Occupied"`
	BookingID *int      `gorm:"column:booking_id" json:"booking_id"`
	Room      Room      `gorm:"foreignKey:RoomNum;references:RoomNum" json:"-"`
	Booking   *Booking  `gorm:"foreignKey:BookingID;references:BookingID" json:"booking,omitempty"`
}

// LastRunning stores the last used running number for ID generation
type LastRunning struct {
	LastRunning int `gorm:"primaryKey;column:last_running" json:"last_running"`
	Year        int `gorm:"column:year" json:"year"`
}
