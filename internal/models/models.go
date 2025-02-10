package models

import "time"

// RoomType has no foreign key dependencies
type RoomType struct {
	TypeID        int    `gorm:"primaryKey;column:type_id" json:"type_id"`
	Name          string `gorm:"unique;not null" json:"name" validate:"required"`
	Description   string `json:"description"`
	Area          int    `gorm:"not null" json:"area" validate:"required"`
	Highlight     string `json:"highlight"`
	Facility      string `json:"facility"`
	PricePerNight int    `gorm:"not null" json:"price_per_night" validate:"required"`
	Capacity      int    `gorm:"not null" json:"capacity" validate:"required"`
	Rooms         []Room `gorm:"foreignKey:TypeID" json:"rooms,omitempty"` // Add this line
}

// Room depends on RoomType
type Room struct {
	RoomNum  int      `gorm:"primaryKey;column:room_num" json:"room_num"`
	TypeID   int      `gorm:"not null" json:"type_id" validate:"required"`
	RoomType RoomType `gorm:"foreignKey:TypeID;references:TypeID" json:"room_type"`
}

// Guest has no foreign key dependencies
type Guest struct {
	GuestID     int       `gorm:"primaryKey;column:guest_id" json:"guest_id"`
	FirstName   string    `gorm:"column:f_name;not null" json:"f_name" validate:"required"`
	LastName    string    `gorm:"column:l_name;not null" json:"l_name" validate:"required"`
	DateOfBirth time.Time `gorm:"not null" json:"date_of_birth" validate:"required"`
	Email       string    `gorm:"unique;not null" json:"email" validate:"required,email"`
	Phone       string    `gorm:"unique;not null" json:"phone" validate:"required"`
	Bookings    []Booking `gorm:"foreignKey:GuestID" json:"bookings,omitempty"` // Add this line
}

// Booking depends on Room and Guest
type Booking struct {
	BookingID    int       `gorm:"primaryKey;column:booking_id" json:"booking_id"`
	RoomNum      int       `gorm:"not null" json:"room_num" validate:"required"`
	GuestID      int       `gorm:"not null" json:"guest_id" validate:"required"`
	CheckInDate  time.Time `gorm:"not null" json:"check_in_date" validate:"required"`
	CheckOutDate time.Time `gorm:"not null" json:"check_out_date" validate:"required"`
	TotalPrice   int       `gorm:"not null" json:"total_price"`
	Room         Room      `gorm:"foreignKey:RoomNum;references:RoomNum" json:"room"`
	Guest        Guest     `gorm:"foreignKey:GuestID" json:"guest"`
}

// RoomStatus depends on Room and Booking
type RoomStatus struct {
	RoomNum   int       `gorm:"primaryKey;column:room_num" json:"room_num"`
	Calendar  time.Time `gorm:"primaryKey;type:date" json:"calendar"`
	Status    string    `gorm:"not null;default:Available" json:"status" validate:"required,oneof=Available Occupied"`
	BookingID *int      `json:"booking_id"`
	Room      Room      `gorm:"foreignKey:RoomNum;references:RoomNum"`
	Booking   *Booking  `gorm:"foreignKey:BookingID;references:BookingID"`
}
