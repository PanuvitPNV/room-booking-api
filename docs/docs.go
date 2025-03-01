// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "termsOfService": "http://swagger.io/terms/",
        "contact": {
            "name": "API Support",
            "url": "http://www.swagger.io/support",
            "email": "support@swagger.io"
        },
        "license": {
            "name": "Apache 2.0",
            "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/bookings": {
            "get": {
                "description": "Find rooms available for booking in a specific date range and guest count",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "bookings"
                ],
                "summary": "Search for available rooms",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Check-in date (YYYY-MM-DD)",
                        "name": "checkIn",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Check-out date (YYYY-MM-DD)",
                        "name": "checkOut",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "description": "Number of guests",
                        "name": "guests",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/models.Room"
                            }
                        }
                    },
                    "400": {
                        "description": "Bad request",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal server error",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        },
        "/bookings/check-availability": {
            "get": {
                "description": "Check if a specific room is available for a date range",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "bookings"
                ],
                "summary": "Check room availability",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "Room number",
                        "name": "roomNum",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Check-in date (YYYY-MM-DD)",
                        "name": "checkIn",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Check-out date (YYYY-MM-DD)",
                        "name": "checkOut",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        },
        "/bookings/with-payment": {
            "post": {
                "description": "Create a new booking with payment in a single atomic transaction. First to pay gets the room.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "bookings"
                ],
                "summary": "Create booking with payment",
                "parameters": [
                    {
                        "description": "Booking and payment details",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handlers.BookingWithPaymentRequest"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/handlers.BookingWithPaymentResponse"
                        }
                    },
                    "400": {
                        "description": "Bad request",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        },
        "/bookings/{id}": {
            "get": {
                "description": "Retrieve details of a specific booking by ID",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "bookings"
                ],
                "summary": "Get booking details",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "Booking ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/models.Booking"
                        }
                    },
                    "400": {
                        "description": "Bad request",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "404": {
                        "description": "Booking not found",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            },
            "put": {
                "description": "Update an existing booking with transaction and concurrency control",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "bookings"
                ],
                "summary": "Update a booking",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "Booking ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Updated booking details",
                        "name": "booking",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/services.BookingRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/models.Booking"
                        }
                    },
                    "400": {
                        "description": "Bad request",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            },
            "delete": {
                "description": "Cancel an existing booking and free up the room",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "bookings"
                ],
                "summary": "Cancel a booking",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "Booking ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Success message",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "400": {
                        "description": "Bad request",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        },
        "/receipts": {
            "post": {
                "description": "Create a payment receipt for a booking with transaction control",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "receipts"
                ],
                "summary": "Create a new receipt",
                "parameters": [
                    {
                        "description": "Receipt details",
                        "name": "receipt",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handlers.ReceiptRequest"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/models.Receipt"
                        }
                    },
                    "400": {
                        "description": "Bad request",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        },
        "/receipts/booking/{bookingId}": {
            "get": {
                "description": "Retrieve a receipt associated with a specific booking",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "receipts"
                ],
                "summary": "Get receipt by booking ID",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "Booking ID",
                        "name": "bookingId",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/models.Receipt"
                        }
                    },
                    "400": {
                        "description": "Bad request",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "404": {
                        "description": "Receipt not found",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        },
        "/receipts/{id}": {
            "get": {
                "description": "Retrieve a receipt by its ID",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "receipts"
                ],
                "summary": "Get receipt by ID",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "Receipt ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/models.Receipt"
                        }
                    },
                    "400": {
                        "description": "Bad request",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "404": {
                        "description": "Receipt not found",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            },
            "put": {
                "description": "Update an existing receipt with transaction control",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "receipts"
                ],
                "summary": "Update a receipt",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "Receipt ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Updated receipt details",
                        "name": "receipt",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handlers.ReceiptRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/models.Receipt"
                        }
                    },
                    "400": {
                        "description": "Bad request",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "404": {
                        "description": "Receipt not found",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            },
            "delete": {
                "description": "Delete an existing receipt with transaction control",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "receipts"
                ],
                "summary": "Delete a receipt",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "Receipt ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Success message",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "400": {
                        "description": "Bad request",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        },
        "/room-types": {
            "get": {
                "description": "Retrieve all room types with their facilities",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "room-types"
                ],
                "summary": "Get all room types",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/models.RoomTypeResponse"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        },
        "/rooms": {
            "get": {
                "description": "Retrieve all hotel rooms with their types and facilities",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "rooms"
                ],
                "summary": "Get all rooms with details",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/models.RoomResponse"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        },
        "/rooms/type/{typeId}": {
            "get": {
                "description": "Retrieve all rooms of a specific room type with facilities",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "rooms"
                ],
                "summary": "Get rooms by type with details",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "Room Type ID",
                        "name": "typeId",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/models.RoomResponse"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        },
        "/rooms/{id}": {
            "get": {
                "description": "Retrieve a specific room with its type and facilities",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "rooms"
                ],
                "summary": "Get a room with details",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "Room Number",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/models.RoomResponse"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        },
        "/rooms/{id}/calendar": {
            "get": {
                "description": "Retrieve the availability calendar for a specific room",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "rooms"
                ],
                "summary": "Get room calendar",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "Room Number",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Start Date (YYYY-MM-DD)",
                        "name": "startDate",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "End Date (YYYY-MM-DD)",
                        "name": "endDate",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/models.RoomStatus"
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "handlers.BookingWithPaymentRequest": {
            "type": "object",
            "properties": {
                "booking": {
                    "$ref": "#/definitions/services.BookingRequest"
                },
                "payment": {
                    "$ref": "#/definitions/services.PaymentRequest"
                }
            }
        },
        "handlers.BookingWithPaymentResponse": {
            "type": "object",
            "properties": {
                "booking": {
                    "$ref": "#/definitions/models.Booking"
                },
                "receipt": {
                    "$ref": "#/definitions/models.Receipt"
                }
            }
        },
        "handlers.ReceiptRequest": {
            "type": "object",
            "required": [
                "amount",
                "booking_id",
                "payment_date",
                "payment_method"
            ],
            "properties": {
                "amount": {
                    "type": "integer",
                    "minimum": 1
                },
                "booking_id": {
                    "type": "integer"
                },
                "payment_date": {
                    "type": "string"
                },
                "payment_method": {
                    "type": "string",
                    "enum": [
                        "Credit",
                        "Debit",
                        "Bank",
                        "Transfer"
                    ]
                }
            }
        },
        "models.Booking": {
            "type": "object",
            "required": [
                "booking_name",
                "check_in_date",
                "check_out_date",
                "room_num"
            ],
            "properties": {
                "booking_date": {
                    "type": "string"
                },
                "booking_id": {
                    "type": "integer"
                },
                "booking_name": {
                    "type": "string"
                },
                "check_in_date": {
                    "type": "string"
                },
                "check_out_date": {
                    "type": "string"
                },
                "receipt": {
                    "$ref": "#/definitions/models.Receipt"
                },
                "room": {
                    "$ref": "#/definitions/models.Room"
                },
                "room_num": {
                    "type": "integer"
                },
                "statuses": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/models.RoomStatus"
                    }
                },
                "total_price": {
                    "type": "integer"
                }
            }
        },
        "models.FacilityResponse": {
            "type": "object",
            "properties": {
                "fac_id": {
                    "type": "integer"
                },
                "name": {
                    "type": "string"
                }
            }
        },
        "models.Receipt": {
            "type": "object",
            "required": [
                "amount",
                "payment_method"
            ],
            "properties": {
                "amount": {
                    "type": "integer"
                },
                "booking_id": {
                    "type": "integer"
                },
                "issue_date": {
                    "type": "string"
                },
                "payment_date": {
                    "type": "string"
                },
                "payment_method": {
                    "type": "string",
                    "enum": [
                        "Credit",
                        "Debit",
                        "Bank",
                        "Transfer"
                    ]
                },
                "receipt_id": {
                    "type": "integer"
                }
            }
        },
        "models.Room": {
            "type": "object",
            "properties": {
                "bookings": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/models.Booking"
                    }
                },
                "room_num": {
                    "type": "integer"
                },
                "room_type": {
                    "$ref": "#/definitions/models.RoomType"
                },
                "statuses": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/models.RoomStatus"
                    }
                },
                "type_id": {
                    "type": "integer"
                }
            }
        },
        "models.RoomFacility": {
            "type": "object",
            "properties": {
                "fac_id": {
                    "type": "integer"
                },
                "type_id": {
                    "type": "integer"
                }
            }
        },
        "models.RoomFacilityResponse": {
            "type": "object",
            "properties": {
                "fac_id": {
                    "type": "integer"
                },
                "facility": {
                    "$ref": "#/definitions/models.FacilityResponse"
                },
                "type_id": {
                    "type": "integer"
                }
            }
        },
        "models.RoomResponse": {
            "type": "object",
            "properties": {
                "room_num": {
                    "type": "integer"
                },
                "room_type": {
                    "$ref": "#/definitions/models.RoomTypeResponse"
                },
                "type_id": {
                    "type": "integer"
                }
            }
        },
        "models.RoomStatus": {
            "type": "object",
            "required": [
                "status"
            ],
            "properties": {
                "booking": {
                    "$ref": "#/definitions/models.Booking"
                },
                "booking_id": {
                    "type": "integer"
                },
                "calendar": {
                    "type": "string"
                },
                "room_num": {
                    "type": "integer"
                },
                "status": {
                    "type": "string",
                    "enum": [
                        "Available",
                        "Occupied"
                    ]
                }
            }
        },
        "models.RoomType": {
            "type": "object",
            "properties": {
                "area": {
                    "type": "integer"
                },
                "description": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "noOfGuest": {
                    "type": "integer"
                },
                "price_per_night": {
                    "type": "integer"
                },
                "room_facilities": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/models.RoomFacility"
                    }
                },
                "rooms": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/models.Room"
                    }
                },
                "type_id": {
                    "type": "integer"
                }
            }
        },
        "models.RoomTypeResponse": {
            "type": "object",
            "properties": {
                "area": {
                    "type": "integer"
                },
                "description": {
                    "type": "string"
                },
                "facilities": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/models.RoomFacilityResponse"
                    }
                },
                "name": {
                    "type": "string"
                },
                "noOfGuest": {
                    "type": "integer"
                },
                "price_per_night": {
                    "type": "integer"
                },
                "type_id": {
                    "type": "integer"
                }
            }
        },
        "services.BookingRequest": {
            "type": "object",
            "required": [
                "booking_name",
                "check_in_date",
                "check_out_date",
                "room_num"
            ],
            "properties": {
                "booking_name": {
                    "type": "string"
                },
                "check_in_date": {
                    "type": "string"
                },
                "check_out_date": {
                    "type": "string"
                },
                "room_num": {
                    "type": "integer"
                }
            }
        },
        "services.PaymentRequest": {
            "type": "object",
            "required": [
                "payment_date",
                "payment_method"
            ],
            "properties": {
                "payment_date": {
                    "type": "string"
                },
                "payment_method": {
                    "type": "string",
                    "enum": [
                        "Credit",
                        "Debit",
                        "Bank",
                        "Transfer"
                    ]
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:8080",
	BasePath:         "/api",
	Schemes:          []string{},
	Title:            "Hotel Booking API",
	Description:      "This is a hotel room booking server with transaction management and concurrency control.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
