// Hotel Booking System Concurrency Test
//
// This program tests the concurrency handling capabilities of the hotel booking API.
// It simulates various real-world scenarios where multiple clients interact with
// the booking system simultaneously, creating contention conditions that stress
// the system's transaction management and concurrency control mechanisms.

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/panuvitpnv/room-booking-api/pkg/report"
)

// Configuration
const (
	DefaultBaseURL = "http://localhost:8080/api/v1"
	NumRooms       = 10
	NumClients     = 20
	TestDuration   = 2 * time.Minute
)

// Available test scenarios
var availableScenarios = []string{
	"peak_booking_rush",
	"weekend_availability_race",
	"payment_processing_surge",
	"cancellation_and_rebooking",
	"booking_modification_conflicts",
	"mixed_clients",
	"all",
}

// Models matching your database schema
type Booking struct {
	BookingID    int       `json:"booking_id,omitempty"`
	BookingName  string    `json:"booking_name"`
	RoomNum      int       `json:"room_num"`
	CheckInDate  time.Time `json:"check_in_date"`
	CheckOutDate time.Time `json:"check_out_date"`
	BookingDate  time.Time `json:"booking_date,omitempty"`
	TotalPrice   int       `json:"total_price,omitempty"`
	Error        string    `json:"error,omitempty"`
	Version      int       `json:"version,omitempty"` // For optimistic concurrency control
}

type Receipt struct {
	ReceiptID     int       `json:"receipt_id,omitempty"`
	BookingID     int       `json:"booking_id"`
	PaymentDate   time.Time `json:"payment_date,omitempty"`
	PaymentMethod string    `json:"payment_method"`
	Amount        int       `json:"amount"`
	IssueDate     time.Time `json:"issue_date,omitempty"`
	Error         string    `json:"error,omitempty"`
}

type Room struct {
	RoomNum int    `json:"room_num"`
	TypeID  int    `json:"type_id"`
	Type    string `json:"type,omitempty"`
}

// TestStatistics tracks test results
type TestStatistics struct {
	mu                  sync.Mutex
	TotalRequests       int
	SuccessfulRequests  int
	FailedRequests      int
	Conflicts           int
	DeadlockPrevention  int
	ConcurrencyIssues   int
	ActionCounts        map[string]int
	ResponseTimes       map[string][]time.Duration
	ErrorsByType        map[string]int
	TransactionsByRoom  map[int]int
	AvailableRoomCounts []int
	TimelineGenerator   *report.TimelineGenerator
}

// ClientAction represents actions a client can take
type ClientAction struct {
	ActionType string
	Data       interface{}
	Result     interface{}
	Error      error
	StartTime  time.Time
	EndTime    time.Time
	ClientID   int // Add this field
}

// Client simulates a user interacting with the booking system
type Client struct {
	ID            int
	BaseURL       string
	Stats         *TestStatistics
	CurrentAction *ClientAction
	BookingIDs    []int
	ReceiptIDs    []int
	Logger        *log.Logger
}

// Initialize statistics
func NewTestStatistics() *TestStatistics {
	return &TestStatistics{
		ActionCounts:        make(map[string]int),
		ResponseTimes:       make(map[string][]time.Duration),
		ErrorsByType:        make(map[string]int),
		TransactionsByRoom:  make(map[int]int),
		AvailableRoomCounts: make([]int, 0),
		TimelineGenerator:   report.NewTimelineGenerator("Hotel Booking System Concurrency Test"),
	}
}

// Update statistics after each action
func (ts *TestStatistics) RecordAction(action *ClientAction) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	// Existing code for recording statistics
	ts.TotalRequests++

	// Record action type
	ts.ActionCounts[action.ActionType]++

	// Record response time
	responseTime := action.EndTime.Sub(action.StartTime)
	ts.ResponseTimes[action.ActionType] = append(ts.ResponseTimes[action.ActionType], responseTime)

	// Extract room number and booking ID from action data if available
	var roomNum int
	var bookingID int

	// Try to extract room and booking information from result or data
	if booking, ok := action.Result.(*Booking); ok && booking != nil {
		roomNum = booking.RoomNum
		bookingID = booking.BookingID
	} else if booking, ok := action.Data.(*Booking); ok && booking != nil {
		roomNum = booking.RoomNum
		bookingID = booking.BookingID
	}

	// Check if action failed
	if action.Error != nil {
		ts.FailedRequests++
		errorType := "other"

		// Categorize errors
		errorMsg := action.Error.Error()
		switch {
		case containsAny(errorMsg, "conflict", "already booked", "unavailable"):
			errorType = "conflict"
			ts.Conflicts++
		case containsAny(errorMsg, "deadlock", "lock", "timeout"):
			errorType = "deadlock_prevented"
			ts.DeadlockPrevention++
		case containsAny(errorMsg, "concurrent", "version", "modified"):
			errorType = "concurrency"
			ts.ConcurrencyIssues++
		}

		ts.ErrorsByType[errorType]++

		// Determine timeline event status
		status := "failed"
		if errorType == "conflict" {
			status = "conflict"
		}

		// Create timeline event for failed action
		event := &report.TransactionEvent{
			ID:           fmt.Sprintf("evt-%d-%s-%d", action.ClientID, action.ActionType, time.Now().UnixNano()),
			ClientID:     action.ClientID,
			ActionType:   action.ActionType,
			RoomNum:      roomNum,
			BookingID:    bookingID,
			StartTime:    action.StartTime,
			EndTime:      action.EndTime,
			Duration:     action.EndTime.Sub(action.StartTime).Milliseconds(),
			Status:       status,
			ErrorMessage: errorMsg,
		}

		// Add event to timeline
		ts.TimelineGenerator.AddEvent(event)
	} else {
		ts.SuccessfulRequests++

		// Track room transactions if this was a booking
		if action.ActionType == "create_booking" {
			if booking, ok := action.Result.(*Booking); ok && booking != nil {
				ts.TransactionsByRoom[booking.RoomNum]++
			}
		}

		// Convert result to JSON string for details
		var details string
		if action.Result != nil {
			if resultJSON, err := json.Marshal(action.Result); err == nil {
				details = string(resultJSON)
			}
		}

		event := &report.TransactionEvent{
			ID:         fmt.Sprintf("evt-%d-%s-%d", action.ClientID, action.ActionType, time.Now().UnixNano()),
			ClientID:   action.ClientID,
			ActionType: action.ActionType,
			RoomNum:    roomNum,
			BookingID:  bookingID,
			StartTime:  action.StartTime,
			EndTime:    action.EndTime,
			Duration:   action.EndTime.Sub(action.StartTime).Milliseconds(),
			Status:     "success",
			Details:    details,
		}

		// Add resource ID if available
		if action.ActionType == "process_payment" {
			if receipt, ok := action.Result.(*Receipt); ok && receipt != nil {
				event.ResourceID = receipt.ReceiptID
			}
		}

		// Add event to timeline
		ts.TimelineGenerator.AddEvent(event)
	}
}

// Create a new client
func NewClient(id int, baseURL string, stats *TestStatistics) *Client {
	return &Client{
		ID:         id,
		BaseURL:    baseURL,
		Stats:      stats,
		BookingIDs: make([]int, 0),
		ReceiptIDs: make([]int, 0),
		Logger:     log.New(os.Stdout, fmt.Sprintf("[Client %d] ", id), log.Ltime),
	}
}

// Helper function to check if a string contains any of the given substrings
func containsAny(s string, substrings ...string) bool {
	for _, substr := range substrings {
		if bytes.Contains([]byte(s), []byte(substr)) {
			return true
		}
	}
	return false
}

// ================== Client Action Methods ==================

// Get available rooms
func (c *Client) GetAvailableRooms(startDate, endDate time.Time) ([]Room, error) {
	c.CurrentAction = &ClientAction{
		ActionType: "get_available_rooms",
		StartTime:  time.Now(),
		ClientID:   c.ID, // Set the client ID
	}
	defer func() {
		c.CurrentAction.EndTime = time.Now()
		c.Stats.RecordAction(c.CurrentAction)
	}()

	request := struct {
		CheckInDate  time.Time `json:"check_in_date"`
		CheckOutDate time.Time `json:"check_out_date"`
	}{
		CheckInDate:  startDate,
		CheckOutDate: endDate,
	}

	var rooms []Room
	err := c.sendRequest("POST", "/rooms/available", request, &rooms)
	if err != nil {
		c.CurrentAction.Error = err
		return nil, err
	}

	c.CurrentAction.Result = rooms
	c.Stats.mu.Lock()
	c.Stats.AvailableRoomCounts = append(c.Stats.AvailableRoomCounts, len(rooms))
	c.Stats.mu.Unlock()

	return rooms, nil
}

// Create a booking
func (c *Client) CreateBooking(name string, roomNum int, checkIn, checkOut time.Time) (*Booking, error) {
	c.CurrentAction = &ClientAction{
		ActionType: "create_booking",
		StartTime:  time.Now(),
		ClientID:   c.ID, // Set the client ID
	}
	defer func() {
		c.CurrentAction.EndTime = time.Now()
		c.Stats.RecordAction(c.CurrentAction)
	}()

	booking := Booking{
		BookingName:  name,
		RoomNum:      roomNum,
		CheckInDate:  checkIn,
		CheckOutDate: checkOut,
	}

	var createdBooking Booking
	err := c.sendRequest("POST", "/bookings", booking, &createdBooking)
	if err != nil {
		c.CurrentAction.Error = err
		return nil, err
	}

	if createdBooking.Error != "" {
		err = fmt.Errorf("booking failed: %s", createdBooking.Error)
		c.CurrentAction.Error = err
		return nil, err
	}

	c.BookingIDs = append(c.BookingIDs, createdBooking.BookingID)
	c.CurrentAction.Result = &createdBooking
	c.Logger.Printf("Created booking %d for room %d from %s to %s",
		createdBooking.BookingID, roomNum,
		checkIn.Format("2006-01-02"), checkOut.Format("2006-01-02"))

	return &createdBooking, nil
}

// Update a booking
func (c *Client) UpdateBooking(bookingID int, newCheckIn, newCheckOut time.Time) error {
	c.CurrentAction = &ClientAction{
		ActionType: "update_booking",
		StartTime:  time.Now(),
		ClientID:   c.ID, // Set the client ID
	}
	defer func() {
		c.CurrentAction.EndTime = time.Now()
		c.Stats.RecordAction(c.CurrentAction)
	}()

	update := struct {
		CheckInDate  time.Time `json:"check_in_date"`
		CheckOutDate time.Time `json:"check_out_date"`
	}{
		CheckInDate:  newCheckIn,
		CheckOutDate: newCheckOut,
	}

	var response map[string]string
	err := c.sendRequest("PUT", fmt.Sprintf("/bookings/%d", bookingID), update, &response)
	if err != nil {
		c.CurrentAction.Error = err
		return err
	}

	if errMsg, exists := response["error"]; exists {
		err = fmt.Errorf("update failed: %s", errMsg)
		c.CurrentAction.Error = err
		return err
	}

	c.CurrentAction.Result = response
	c.Logger.Printf("Updated booking %d with new dates: %s to %s",
		bookingID, newCheckIn.Format("2006-01-02"), newCheckOut.Format("2006-01-02"))

	return nil
}

// Process payment for a booking
func (c *Client) ProcessPayment(bookingID, amount int, method string) (*Receipt, error) {
	c.CurrentAction = &ClientAction{
		ActionType: "process_payment",
		StartTime:  time.Now(),
		ClientID:   c.ID, // Set the client ID
	}
	defer func() {
		c.CurrentAction.EndTime = time.Now()
		c.Stats.RecordAction(c.CurrentAction)
	}()

	payment := struct {
		BookingID     int    `json:"booking_id"`
		PaymentMethod string `json:"payment_method"`
		Amount        int    `json:"amount"`
	}{
		BookingID:     bookingID,
		PaymentMethod: method,
		Amount:        amount,
	}

	var receipt Receipt
	err := c.sendRequest("POST", "/receipts", payment, &receipt)
	if err != nil {
		c.CurrentAction.Error = err
		return nil, err
	}

	if receipt.Error != "" {
		err = fmt.Errorf("payment failed: %s", receipt.Error)
		c.CurrentAction.Error = err
		return nil, err
	}

	c.ReceiptIDs = append(c.ReceiptIDs, receipt.ReceiptID)
	c.CurrentAction.Result = &receipt
	c.Logger.Printf("Processed payment for booking %d, amount: $%d, receipt ID: %d",
		bookingID, amount, receipt.ReceiptID)

	return &receipt, nil
}

// Cancel a booking
func (c *Client) CancelBooking(bookingID int) error {
	c.CurrentAction = &ClientAction{
		ActionType: "cancel_booking",
		StartTime:  time.Now(),
		ClientID:   c.ID, // Set the client ID
	}
	defer func() {
		c.CurrentAction.EndTime = time.Now()
		c.Stats.RecordAction(c.CurrentAction)
	}()

	var response map[string]string
	err := c.sendRequest("DELETE", fmt.Sprintf("/bookings/%d", bookingID), nil, &response)
	if err != nil {
		c.CurrentAction.Error = err
		return err
	}

	if errMsg, exists := response["error"]; exists {
		err = fmt.Errorf("cancellation failed: %s", errMsg)
		c.CurrentAction.Error = err
		return err
	}

	c.CurrentAction.Result = response
	c.Logger.Printf("Cancelled booking %d", bookingID)

	// Remove from client's booking list
	for i, id := range c.BookingIDs {
		if id == bookingID {
			c.BookingIDs = append(c.BookingIDs[:i], c.BookingIDs[i+1:]...)
			break
		}
	}

	return nil
}

// Process a refund
func (c *Client) ProcessRefund(bookingID int) error {
	c.CurrentAction = &ClientAction{
		ActionType: "process_refund",
		StartTime:  time.Now(),
		ClientID:   c.ID, // Set the client ID
	}
	defer func() {
		c.CurrentAction.EndTime = time.Now()
		c.Stats.RecordAction(c.CurrentAction)
	}()

	refundRequest := struct {
		BookingID int `json:"booking_id"`
	}{
		BookingID: bookingID,
	}

	var response map[string]string
	err := c.sendRequest("POST", "/receipts/refund", refundRequest, &response)
	if err != nil {
		c.CurrentAction.Error = err
		return err
	}

	if errMsg, exists := response["error"]; exists {
		err = fmt.Errorf("refund failed: %s", errMsg)
		c.CurrentAction.Error = err
		return err
	}

	c.CurrentAction.Result = response
	c.Logger.Printf("Processed refund for booking %d", bookingID)

	return nil
}

// Get a booking by ID
func (c *Client) GetBooking(bookingID int) (*Booking, error) {
	c.CurrentAction = &ClientAction{
		ActionType: "get_booking",
		StartTime:  time.Now(),
		ClientID:   c.ID, // Set the client ID
	}
	defer func() {
		c.CurrentAction.EndTime = time.Now()
		c.Stats.RecordAction(c.CurrentAction)
	}()

	var booking Booking
	err := c.sendRequest("GET", fmt.Sprintf("/bookings/%d", bookingID), nil, &booking)
	if err != nil {
		c.CurrentAction.Error = err
		return nil, err
	}

	c.CurrentAction.Result = &booking
	return &booking, nil
}

// Helper method to send HTTP requests
func (c *Client) sendRequest(method, endpoint string, body interface{}, response interface{}) error {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("error marshaling request: %v", err)
		}
	}

	url := c.BaseURL + endpoint
	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Add a random identifier to help trace concurrent requests
	req.Header.Set("X-Client-ID", fmt.Sprintf("client-%d-%d", c.ID, rand.Intn(1000)))

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errorResp map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			return fmt.Errorf("HTTP error: %d", resp.StatusCode)
		}
		if errMsg, exists := errorResp["error"]; exists {
			return fmt.Errorf("API error: %s", errMsg)
		}
		return fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	if response != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
			return fmt.Errorf("error decoding response: %v", err)
		}
	}

	return nil
}

// ================== Scenario Generation ==================

// Generate random date in the future
func randomFutureDate(minDays, maxDays int) time.Time {
	daysToAdd := minDays + rand.Intn(maxDays-minDays+1)
	return time.Now().AddDate(0, 0, daysToAdd)
}

// Generate a random stay duration
func randomStayDuration() (time.Time, time.Time) {
	checkIn := randomFutureDate(1, 60)
	stayLength := 1 + rand.Intn(5) // 1-5 days
	checkOut := checkIn.AddDate(0, 0, stayLength)
	return checkIn, checkOut
}

// Random payment method
func randomPaymentMethod() string {
	methods := []string{"Credit", "Debit", "Bank Transfer"}
	return methods[rand.Intn(len(methods))]
}

// ================== Client Behavior ==================

// Run a client that performs a sequence of actions
func (c *Client) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	// Simulate different client behavior patterns
	switch c.ID % 5 {
	case 0:
		// Normal booker: books, pays, occasionally updates or cancels
		c.normalBookerBehavior(ctx)
	case 1:
		// Indecisive customer: creates multiple bookings, updates frequently, cancels some
		c.indecisiveCustomerBehavior(ctx)
	case 2:
		// Group coordinator: books multiple rooms for same dates
		c.groupCoordinatorBehavior(ctx)
	case 3:
		// Last-minute booker: books rooms with very near check-in dates
		c.lastMinuteBookerBehavior(ctx)
	case 4:
		// Premium customer: always books specific room types, pays immediately
		c.premiumCustomerBehavior(ctx)
	}
}

// Normal booker behavior
func (c *Client) normalBookerBehavior(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Check available rooms
			checkIn, checkOut := randomStayDuration()
			rooms, err := c.GetAvailableRooms(checkIn, checkOut)
			if err != nil || len(rooms) == 0 {
				time.Sleep(500 * time.Millisecond)
				continue
			}

			// Book a random available room
			roomIndex := rand.Intn(len(rooms))
			room := rooms[roomIndex]
			booking, err := c.CreateBooking(
				fmt.Sprintf("Normal Customer %d", c.ID),
				room.RoomNum,
				checkIn,
				checkOut,
			)
			if err != nil {
				time.Sleep(500 * time.Millisecond)
				continue
			}

			// Process payment (80% of the time)
			if rand.Float32() < 0.8 {
				_, err = c.ProcessPayment(booking.BookingID, booking.TotalPrice, randomPaymentMethod())
				if err != nil {
					// Try to cancel if payment fails
					c.CancelBooking(booking.BookingID)
					time.Sleep(500 * time.Millisecond)
					continue
				}
			}

			// Maybe update booking (30% chance)
			if rand.Float32() < 0.3 {
				// Modify check-out date
				newCheckOut := checkOut.AddDate(0, 0, 1+rand.Intn(2)) // Extend by 1-2 days
				c.UpdateBooking(booking.BookingID, checkIn, newCheckOut)
			}

			// Maybe cancel (10% chance)
			if rand.Float32() < 0.1 {
				c.CancelBooking(booking.BookingID)
				// If cancelled, try to get a refund
				c.ProcessRefund(booking.BookingID)
			}

			// Sleep before next action
			time.Sleep(time.Duration(1+rand.Intn(3)) * time.Second)
		}
	}
}

// Indecisive customer behavior
func (c *Client) indecisiveCustomerBehavior(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Create 2-3 different bookings
			var bookings []*Booking
			numBookings := 2 + rand.Intn(2)

			for i := 0; i < numBookings; i++ {
				checkIn, checkOut := randomStayDuration()
				rooms, err := c.GetAvailableRooms(checkIn, checkOut)
				if err != nil || len(rooms) == 0 {
					continue
				}

				room := rooms[rand.Intn(len(rooms))]
				booking, err := c.CreateBooking(
					fmt.Sprintf("Indecisive Customer %d", c.ID),
					room.RoomNum,
					checkIn,
					checkOut,
				)
				if err == nil {
					bookings = append(bookings, booking)
				}
			}

			// Update bookings multiple times
			for _, booking := range bookings {
				for i := 0; i < 1+rand.Intn(3); i++ { // 1-3 updates
					// Either change check-in or check-out
					var newCheckIn, newCheckOut time.Time
					if rand.Float32() < 0.5 {
						// Modify check-in (earlier or later by 1-2 days)
						adjustment := -2 + rand.Intn(5) // -2 to +2 days
						newCheckIn = booking.CheckInDate.AddDate(0, 0, adjustment)
						newCheckOut = booking.CheckOutDate
					} else {
						// Modify check-out (shorter or longer stay by 1-2 days)
						adjustment := -2 + rand.Intn(5) // -2 to +2 days
						newCheckIn = booking.CheckInDate
						newCheckOut = booking.CheckOutDate.AddDate(0, 0, adjustment)
					}

					// Only proceed if dates are valid (check-in before check-out)
					if newCheckIn.Before(newCheckOut) {
						c.UpdateBooking(booking.BookingID, newCheckIn, newCheckOut)
					}

					time.Sleep(200 * time.Millisecond)
				}
			}

			// Keep one booking, cancel the rest
			if len(bookings) > 0 {
				keepIndex := rand.Intn(len(bookings))

				for i, booking := range bookings {
					if i != keepIndex {
						c.CancelBooking(booking.BookingID)
					} else {
						// Process payment for the booking we're keeping
						c.ProcessPayment(booking.BookingID, booking.TotalPrice, randomPaymentMethod())
					}
				}
			}

			// Sleep longer between cycles
			time.Sleep(time.Duration(3+rand.Intn(5)) * time.Second)
		}
	}
}

// Group coordinator behavior
func (c *Client) groupCoordinatorBehavior(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Book multiple rooms for the same dates (like for a group trip)
			checkIn, checkOut := randomStayDuration()
			rooms, err := c.GetAvailableRooms(checkIn, checkOut)
			if err != nil || len(rooms) < 2 {
				time.Sleep(500 * time.Millisecond)
				continue
			}

			// Try to book 2-4 rooms
			numRoomsToBook := 2 + rand.Intn(3)
			if numRoomsToBook > len(rooms) {
				numRoomsToBook = len(rooms)
			}

			var bookedRooms []*Booking
			for i := 0; i < numRoomsToBook; i++ {
				room := rooms[i]
				guestName := fmt.Sprintf("Group Member %d-%d", c.ID, i+1)
				booking, err := c.CreateBooking(guestName, room.RoomNum, checkIn, checkOut)
				if err == nil {
					bookedRooms = append(bookedRooms, booking)
				}
			}

			// Process payments for all bookings
			for _, booking := range bookedRooms {
				c.ProcessPayment(booking.BookingID, booking.TotalPrice, randomPaymentMethod())
				time.Sleep(100 * time.Millisecond)
			}

			// Occasionally one person from the group cancels (20% chance)
			if len(bookedRooms) > 0 && rand.Float32() < 0.2 {
				cancelIndex := rand.Intn(len(bookedRooms))
				booking := bookedRooms[cancelIndex]
				c.CancelBooking(booking.BookingID)
				c.ProcessRefund(booking.BookingID)
			}

			// Sleep longer between group bookings
			time.Sleep(time.Duration(5+rand.Intn(10)) * time.Second)
		}
	}
}

// Last-minute booker behavior
func (c *Client) lastMinuteBookerBehavior(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Book rooms with very near check-in dates
			checkIn := randomFutureDate(0, 3) // 0-3 days in the future
			stayLength := 1 + rand.Intn(3)    // 1-3 day stay
			checkOut := checkIn.AddDate(0, 0, stayLength)

			rooms, err := c.GetAvailableRooms(checkIn, checkOut)
			if err != nil || len(rooms) == 0 {
				time.Sleep(200 * time.Millisecond)
				continue
			}

			// Book a random available room
			room := rooms[rand.Intn(len(rooms))]
			booking, err := c.CreateBooking(
				fmt.Sprintf("Last-Minute Guest %d", c.ID),
				room.RoomNum,
				checkIn,
				checkOut,
			)
			if err != nil {
				time.Sleep(200 * time.Millisecond)
				continue
			}

			// Always pay immediately
			c.ProcessPayment(booking.BookingID, booking.TotalPrice, randomPaymentMethod())

			// Sleep shorter between actions (more frequent)
			time.Sleep(time.Duration(1+rand.Intn(2)) * time.Second)
		}
	}
}

// Premium customer behavior
func (c *Client) premiumCustomerBehavior(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Premium customers look further ahead
			checkIn := randomFutureDate(10, 120) // 10-120 days in advance
			stayLength := 3 + rand.Intn(5)       // 3-7 day stay
			checkOut := checkIn.AddDate(0, 0, stayLength)

			rooms, err := c.GetAvailableRooms(checkIn, checkOut)
			if err != nil || len(rooms) == 0 {
				time.Sleep(500 * time.Millisecond)
				continue
			}

			// Premium customers prefer specific room types (try higher room numbers)
			// In this simulation, we'll just pick rooms with higher numbers as a proxy
			var selectedRoom Room
			if len(rooms) >= 3 {
				// Sort by room number (simplified approach)
				maxRoomNum := 0
				for _, room := range rooms {
					if room.RoomNum > maxRoomNum {
						maxRoomNum = room.RoomNum
						selectedRoom = room
					}
				}
			} else {
				selectedRoom = rooms[rand.Intn(len(rooms))]
			}

			booking, err := c.CreateBooking(
				fmt.Sprintf("Premium Guest %d", c.ID),
				selectedRoom.RoomNum,
				checkIn,
				checkOut,
			)
			if err != nil {
				time.Sleep(500 * time.Millisecond)
				continue
			}

			// Always pay immediately
			c.ProcessPayment(booking.BookingID, booking.TotalPrice, "Credit") // Premium customers use credit

			// Less likely to cancel (5% chance)
			if rand.Float32() < 0.05 {
				c.CancelBooking(booking.BookingID)
				c.ProcessRefund(booking.BookingID)
			}

			// More likely to modify stay (extend)
			if rand.Float32() < 0.4 {
				// Usually extend stay
				newCheckOut := checkOut.AddDate(0, 0, 1+rand.Intn(3)) // Extend by 1-3 days
				c.UpdateBooking(booking.BookingID, checkIn, newCheckOut)

				// And pay the difference (simplified - in real app would calculate)
				c.ProcessPayment(booking.BookingID, 200+rand.Intn(300), "Credit")
			}

			// Sleep longer between bookings
			time.Sleep(time.Duration(3+rand.Intn(5)) * time.Second)
		}
	}
}

// ================== Concurrent Scenarios ==================

// Run a specific concurrent booking scenario
func runConcurrentBookingScenario(baseURL string, scenario string) {
	log.Printf("Running concurrent scenario: %s", scenario)
	stats := NewTestStatistics()

	switch scenario {
	case "peak_booking_rush":
		runPeakBookingRush(baseURL, stats)
	case "weekend_availability_race":
		runWeekendAvailabilityRace(baseURL, stats)
	case "payment_processing_surge":
		runPaymentProcessingSurge(baseURL, stats)
	case "cancellation_and_rebooking":
		runCancellationAndRebooking(baseURL, stats)
	case "booking_modification_conflicts":
		runBookingModificationConflicts(baseURL, stats)
	}

	// Print scenario statistics
	printScenarioStatistics(scenario, stats)
}

// Print runtime statistics during test execution
func printRuntimeStatistics(stats *TestStatistics) {
	stats.mu.Lock()
	defer stats.mu.Unlock()

	fmt.Println("\n--- RUNTIME STATISTICS ---")
	fmt.Printf("Requests: %d (Success: %d, Failed: %d)\n",
		stats.TotalRequests, stats.SuccessfulRequests, stats.FailedRequests)
	fmt.Printf("Concurrency events: %d conflicts, %d deadlock prevented\n",
		stats.Conflicts, stats.DeadlockPrevention)

	// Print top 3 actions
	type actionCount struct {
		action string
		count  int
	}

	actions := make([]actionCount, 0, len(stats.ActionCounts))
	for action, count := range stats.ActionCounts {
		actions = append(actions, actionCount{action, count})
	}

	sort.Slice(actions, func(i, j int) bool {
		return actions[i].count > actions[j].count
	})

	fmt.Println("Top actions:")
	for i := 0; i < min(3, len(actions)); i++ {
		fmt.Printf("- %s: %d\n", actions[i].action, actions[i].count)
	}
	fmt.Println("-------------------------")
}

// Main function - entry point for the test program
func main() {
	// Set up randomization
	rand.Seed(time.Now().UnixNano())

	// Parse command line flags
	baseURL := DefaultBaseURL
	if len(os.Args) > 1 {
		baseURL = os.Args[1]
	}

	scenario := "mixed_clients"
	if len(os.Args) > 2 {
		scenario = os.Args[2]
	}

	// Print header
	fmt.Println("==================================================")
	fmt.Println("  HOTEL BOOKING SYSTEM CONCURRENCY TEST")
	fmt.Println("==================================================")
	fmt.Printf("Base URL: %s\n", baseURL)
	fmt.Printf("Test Scenario: %s\n", scenario)
	fmt.Println("==================================================")

	// Run the specified scenario
	if scenario == "all" {
		// Run all scenarios in sequence
		for _, s := range availableScenarios {
			if s != "all" && s != "mixed_clients" {
				runConcurrentBookingScenario(baseURL, s)
				time.Sleep(2 * time.Second) // Brief pause between scenarios
			}
		}

		// Run the mixed client scenario last
		fmt.Println("\nRunning final mixed client scenario...")
		runMixedClientScenario(baseURL, NumClients, TestDuration)
	} else if scenario == "mixed_clients" {
		// Run just the mixed scenario
		runMixedClientScenario(baseURL, NumClients, TestDuration)
	} else {
		// Run a specific scenario
		runConcurrentBookingScenario(baseURL, scenario)
	}

	fmt.Println("\n==================================================")
	fmt.Println("  TEST EXECUTION COMPLETE")
	fmt.Println("==================================================")
}

// Print statistics for a completed scenario
func printScenarioStatistics(scenario string, stats *TestStatistics) {
	fmt.Println("\n--------------------------------------")
	fmt.Printf("RESULTS FOR SCENARIO: %s\n", strings.ToUpper(scenario))
	fmt.Println("--------------------------------------")
	fmt.Printf("Total Requests: %d\n", stats.TotalRequests)
	fmt.Printf("Successful: %d (%.1f%%)\n",
		stats.SuccessfulRequests,
		float64(stats.SuccessfulRequests)/float64(stats.TotalRequests)*100)
	fmt.Printf("Failed: %d (%.1f%%)\n",
		stats.FailedRequests,
		float64(stats.FailedRequests)/float64(stats.TotalRequests)*100)

	fmt.Println("\nCONCURRENCY STATISTICS:")
	fmt.Printf("Conflicts Detected: %d\n", stats.Conflicts)
	fmt.Printf("Deadlock Prevention: %d\n", stats.DeadlockPrevention)
	fmt.Printf("Optimistic Concurrency Issues: %d\n", stats.ConcurrencyIssues)

	fmt.Println("\nACTIONS BY TYPE:")
	for action, count := range stats.ActionCounts {
		fmt.Printf("- %s: %d\n", action, count)
	}

	fmt.Println("\nAVERAGE RESPONSE TIMES:")
	for action, times := range stats.ResponseTimes {
		if len(times) > 0 {
			var total time.Duration
			for _, t := range times {
				total += t
			}
			avg := total / time.Duration(len(times))
			fmt.Printf("- %s: %v\n", action, avg)
		}
	}

	if len(stats.ErrorsByType) > 0 {
		fmt.Println("\nERROR TYPES:")
		for errType, count := range stats.ErrorsByType {
			fmt.Printf("- %s: %d\n", errType, count)
		}
	}

	fmt.Println("\nROOM STATISTICS:")
	// Print the most contended rooms
	type roomContention struct {
		roomNum int
		count   int
	}

	roomStats := make([]roomContention, 0, len(stats.TransactionsByRoom))
	for room, count := range stats.TransactionsByRoom {
		roomStats = append(roomStats, roomContention{room, count})
	}

	// Sort by contention (highest first)
	sort.Slice(roomStats, func(i, j int) bool {
		return roomStats[i].count > roomStats[j].count
	})

	// Print top 5 most contended rooms
	fmt.Println("Most Contended Rooms:")
	for i := 0; i < min(5, len(roomStats)); i++ {
		fmt.Printf("- Room %d: %d transactions\n",
			roomStats[i].roomNum, roomStats[i].count)
	}

	if len(stats.AvailableRoomCounts) > 0 {
		// Calculate average available rooms
		total := 0
		for _, count := range stats.AvailableRoomCounts {
			total += count
		}
		avg := float64(total) / float64(len(stats.AvailableRoomCounts))
		fmt.Printf("\nAverage Available Rooms: %.1f\n", avg)
	}

	fmt.Println("--------------------------------------")
	generateTimelineReport(stats, scenario)
}

// Peak booking rush scenario
func runPeakBookingRush(baseURL string, stats *TestStatistics) {
	// Simulate a rush of clients all trying to book during a peak period
	numClients := 15

	// Target period (e.g., New Year's Eve)
	peakCheckIn := time.Date(time.Now().Year(), 12, 30, 0, 0, 0, 0, time.Local)
	peakCheckOut := time.Date(time.Now().Year()+1, 1, 2, 0, 0, 0, 0, time.Local)

	// Get list of all rooms first
	client := NewClient(0, baseURL, stats)
	rooms, err := client.GetAvailableRooms(peakCheckIn, peakCheckOut)
	if err != nil || len(rooms) == 0 {
		log.Printf("Error getting rooms for peak period: %v", err)
		return
	}

	log.Printf("Found %d available rooms for peak period", len(rooms))

	// Create a pool of clients all trying to book at once
	var wg sync.WaitGroup
	wg.Add(numClients)

	for i := 1; i <= numClients; i++ {
		go func(clientID int) {
			defer wg.Done()
			client := NewClient(clientID, baseURL, stats)

			// Each client tries to book randomly from the same room pool
			roomIndex := rand.Intn(len(rooms))
			room := rooms[roomIndex]

			booking, err := client.CreateBooking(
				fmt.Sprintf("Peak Rush Customer %d", clientID),
				room.RoomNum,
				peakCheckIn,
				peakCheckOut,
			)

			// If booking succeeds, try to process payment right away
			if err == nil && booking != nil {
				client.ProcessPayment(booking.BookingID, booking.TotalPrice, randomPaymentMethod())
			}
		}(i)
	}

	// Wait for all bookings to complete
	wg.Wait()
	log.Printf("Peak booking rush scenario completed")
}

// Weekend availability race scenario
func runWeekendAvailabilityRace(baseURL string, stats *TestStatistics) {
	// Find next weekend
	now := time.Now()
	daysUntilFriday := (5 - int(now.Weekday()) + 7) % 7
	if daysUntilFriday == 0 {
		daysUntilFriday = 7 // Next Friday, not today
	}

	fridayCheckIn := now.AddDate(0, 0, daysUntilFriday)
	sundayCheckOut := fridayCheckIn.AddDate(0, 0, 2)

	// Format dates properly
	checkIn := time.Date(fridayCheckIn.Year(), fridayCheckIn.Month(), fridayCheckIn.Day(), 15, 0, 0, 0, time.Local)
	checkOut := time.Date(sundayCheckOut.Year(), sundayCheckOut.Month(), sundayCheckOut.Day(), 11, 0, 0, 0, time.Local)

	log.Printf("Weekend race scenario for %s to %s",
		checkIn.Format("Mon Jan 2"), checkOut.Format("Mon Jan 2"))

	// Create two phases: first check availability, then rush to book
	// This simulates real-world behavior of users first checking then quickly trying to book

	// Phase 1: Multiple clients check availability
	numClients := 12
	var availableRooms []Room

	// All clients check availability first
	client := NewClient(0, baseURL, stats)
	rooms, err := client.GetAvailableRooms(checkIn, checkOut)
	if err != nil || len(rooms) == 0 {
		log.Printf("Error getting weekend rooms: %v", err)
		return
	}

	availableRooms = rooms
	log.Printf("Found %d available rooms for weekend", len(availableRooms))

	// Phase 2: All clients rush to book at once
	var wg sync.WaitGroup
	wg.Add(numClients)

	// Create a pool of booking clients
	for i := 1; i <= numClients; i++ {
		go func(clientID int) {
			defer wg.Done()
			client := NewClient(clientID, baseURL, stats)

			// Each client focuses on booking a specific room - creating contention
			roomIndex := clientID % len(availableRooms)
			targetRoom := availableRooms[roomIndex]

			// Small staggered delay to make it more realistic
			time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)

			booking, err := client.CreateBooking(
				fmt.Sprintf("Weekend Traveler %d", clientID),
				targetRoom.RoomNum,
				checkIn,
				checkOut,
			)

			if err == nil && booking != nil {
				// Successfully booked, process payment
				client.ProcessPayment(booking.BookingID, booking.TotalPrice, randomPaymentMethod())

				// 20% chance to cancel and rebook a different room
				if rand.Float32() < 0.2 {
					client.CancelBooking(booking.BookingID)

					// Try booking a different room
					if len(availableRooms) > 1 {
						newRoomIndex := (roomIndex + 1) % len(availableRooms)
						newRoom := availableRooms[newRoomIndex]

						newBooking, err := client.CreateBooking(
							fmt.Sprintf("Weekend Traveler %d", clientID),
							newRoom.RoomNum,
							checkIn,
							checkOut,
						)

						if err == nil && newBooking != nil {
							client.ProcessPayment(newBooking.BookingID, newBooking.TotalPrice, randomPaymentMethod())
						}
					}
				}
			}
		}(i)
	}

	wg.Wait()
	log.Printf("Weekend availability race scenario completed")
}

// Payment processing surge scenario
func runPaymentProcessingSurge(baseURL string, stats *TestStatistics) {
	// Create a set of bookings first, then process payments concurrently
	numBookings := 10
	bookings := make([]*Booking, 0, numBookings)

	// Create the bookings first (sequential)
	setupClient := NewClient(0, baseURL, stats)
	for i := 0; i < numBookings; i++ {
		checkIn, checkOut := randomStayDuration()

		rooms, err := setupClient.GetAvailableRooms(checkIn, checkOut)
		if err != nil || len(rooms) == 0 {
			continue // Skip if no rooms available
		}

		room := rooms[rand.Intn(len(rooms))]
		booking, err := setupClient.CreateBooking(
			fmt.Sprintf("Payment Test Guest %d", i),
			room.RoomNum,
			checkIn,
			checkOut,
		)

		if err == nil && booking != nil {
			bookings = append(bookings, booking)
		}

		// Small delay between bookings
		time.Sleep(200 * time.Millisecond)
	}

	log.Printf("Created %d bookings for payment test", len(bookings))
	if len(bookings) == 0 {
		return
	}

	// Now process payments concurrently, simulating end-of-day batch processing
	var wg sync.WaitGroup
	numPaymentClients := len(bookings) * 2 // More clients than bookings to create contention

	// Create a random subset of duplicated booking IDs to simulate multiple payment attempts
	// This creates situations where the same booking might be paid for twice
	paymentBookings := make([]int, numPaymentClients)
	for i := 0; i < numPaymentClients; i++ {
		// Intentionally create some duplicate payment attempts
		bookingIndex := rand.Intn(len(bookings))
		paymentBookings[i] = bookings[bookingIndex].BookingID
	}

	// Shuffle to randomize the order
	rand.Shuffle(len(paymentBookings), func(i, j int) {
		paymentBookings[i], paymentBookings[j] = paymentBookings[j], paymentBookings[i]
	})

	// Launch concurrent payment processing
	wg.Add(numPaymentClients)
	for i := 0; i < numPaymentClients; i++ {
		go func(clientID int, bookingID int) {
			defer wg.Done()

			client := NewClient(clientID, baseURL, stats)

			// Find the booking details to get the correct amount
			var bookingAmount int
			for _, b := range bookings {
				if b.BookingID == bookingID {
					bookingAmount = b.TotalPrice
					break
				}
			}

			if bookingAmount == 0 {
				bookingAmount = 1000 // Fallback if not found
			}

			// Process the payment
			_, err := client.ProcessPayment(
				bookingID,
				bookingAmount,
				randomPaymentMethod(),
			)

			if err != nil {
				log.Printf("Client %d payment failed for booking %d: %v",
					clientID, bookingID, err)
			}
		}(i, paymentBookings[i])
	}

	wg.Wait()
	log.Printf("Payment processing surge scenario completed")
}

// Cancellation and rebooking scenario
func runCancellationAndRebooking(baseURL string, stats *TestStatistics) {
	// This scenario simulates a cancellation creating an opportunity that multiple
	// clients try to take advantage of simultaneously

	// Step 1: Book all rooms for a specific date range
	checkIn := time.Now().AddDate(0, 0, 14) // Two weeks from now
	checkOut := checkIn.AddDate(0, 0, 3)    // 3-day stay

	// Format dates properly
	checkIn = time.Date(checkIn.Year(), checkIn.Month(), checkIn.Day(), 15, 0, 0, 0, time.Local)
	checkOut = time.Date(checkOut.Year(), checkOut.Month(), checkOut.Day(), 11, 0, 0, 0, time.Local)

	// Set up initial client
	setupClient := NewClient(0, baseURL, stats)

	// Get all available rooms
	availableRooms, err := setupClient.GetAvailableRooms(checkIn, checkOut)
	if err != nil || len(availableRooms) == 0 {
		log.Printf("Error getting rooms: %v", err)
		return
	}

	// Book some rooms to create initial scarcity
	initialBookings := make([]*Booking, 0)
	roomsToBook := int(float64(len(availableRooms)) * 0.7) // Book 70% of available rooms

	for i := 0; i < roomsToBook && i < len(availableRooms); i++ {
		booking, err := setupClient.CreateBooking(
			fmt.Sprintf("Initial Booker %d", i),
			availableRooms[i].RoomNum,
			checkIn,
			checkOut,
		)

		if err == nil && booking != nil {
			initialBookings = append(initialBookings, booking)
			setupClient.ProcessPayment(booking.BookingID, booking.TotalPrice, randomPaymentMethod())
		}

		time.Sleep(100 * time.Millisecond)
	}

	log.Printf("Created %d initial bookings out of %d rooms",
		len(initialBookings), len(availableRooms))

	if len(initialBookings) == 0 {
		return
	}

	// Step 2: Cancel one booking to create an opportunity
	cancelIndex := rand.Intn(len(initialBookings))
	bookingToCancel := initialBookings[cancelIndex]

	log.Printf("Cancelling booking %d for room %d",
		bookingToCancel.BookingID, bookingToCancel.RoomNum)

	err = setupClient.CancelBooking(bookingToCancel.BookingID)
	if err != nil {
		log.Printf("Error cancelling booking: %v", err)
		return
	}

	// Step 3: Have multiple clients race to book the newly available room
	cancelledRoom := bookingToCancel.RoomNum
	numRebookers := 8

	var wg sync.WaitGroup
	wg.Add(numRebookers)

	log.Printf("Launching %d clients to try booking room %d",
		numRebookers, cancelledRoom)

	for i := 1; i <= numRebookers; i++ {
		go func(clientID int) {
			defer wg.Done()

			client := NewClient(clientID, baseURL, stats)

			// Add a small random delay to simulate network variability
			time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)

			// Try to book the cancelled room
			booking, err := client.CreateBooking(
				fmt.Sprintf("Rebooker %d", clientID),
				cancelledRoom,
				checkIn,
				checkOut,
			)

			if err == nil && booking != nil {
				log.Printf("Client %d successfully rebooked room %d",
					clientID, cancelledRoom)

				// Pay for successful booking
				client.ProcessPayment(booking.BookingID, booking.TotalPrice, randomPaymentMethod())

				// 30% chance to modify the booking
				if rand.Float32() < 0.3 {
					// Extend or shorten stay by 1 day
					adjustment := []int{-1, 1}[rand.Intn(2)]
					newCheckOut := checkOut.AddDate(0, 0, adjustment)

					// Only adjust if it makes sense
					if newCheckOut.After(checkIn) {
						client.UpdateBooking(booking.BookingID, checkIn, newCheckOut)
					}
				}
			}
		}(i)
	}

	wg.Wait()
	log.Printf("Cancellation and rebooking scenario completed")
}

// Booking modification conflicts scenario
func runBookingModificationConflicts(baseURL string, stats *TestStatistics) {
	// This scenario tests concurrent modifications to the same booking
	// Simulates situations like multiple agents working on same booking or
	// a customer using multiple devices/tabs

	// Create a booking to work with
	setupClient := NewClient(0, baseURL, stats)
	checkIn := time.Now().AddDate(0, 1, 0) // One month from now
	checkOut := checkIn.AddDate(0, 0, 5)   // 5-day stay

	// Format dates properly
	checkIn = time.Date(checkIn.Year(), checkIn.Month(), checkIn.Day(), 15, 0, 0, 0, time.Local)
	checkOut = time.Date(checkOut.Year(), checkOut.Month(), checkOut.Day(), 11, 0, 0, 0, time.Local)

	// Get a room to book
	rooms, err := setupClient.GetAvailableRooms(checkIn, checkOut)
	if err != nil || len(rooms) == 0 {
		log.Printf("Error getting rooms for modification test: %v", err)
		return
	}

	// Create the initial booking
	room := rooms[rand.Intn(len(rooms))]
	booking, err := setupClient.CreateBooking(
		"Modification Test Guest",
		room.RoomNum,
		checkIn,
		checkOut,
	)

	if err != nil || booking == nil {
		log.Printf("Failed to create test booking: %v", err)
		return
	}

	log.Printf("Created test booking %d for modification conflicts", booking.BookingID)

	// Process payment for the booking
	_, err = setupClient.ProcessPayment(booking.BookingID, booking.TotalPrice, "Credit")
	if err != nil {
		log.Printf("Warning: Payment processing failed: %v", err)
		// Continue anyway for this test
	}

	// Now create multiple "clients" that will try to modify the same booking
	numModifiers := 5
	var wg sync.WaitGroup
	wg.Add(numModifiers)

	for i := 1; i <= numModifiers; i++ {
		go func(clientID int) {
			defer wg.Done()

			client := NewClient(clientID, baseURL, stats)

			// Each client tries a different type of modification
			var newCheckIn, newCheckOut time.Time

			switch clientID % 5 {
			case 0:
				// Extend stay by 2 days
				newCheckIn = checkIn
				newCheckOut = checkOut.AddDate(0, 0, 2)
			case 1:
				// Shorten stay by 1 day
				newCheckIn = checkIn
				newCheckOut = checkOut.AddDate(0, 0, -1)
			case 2:
				// Arrive 1 day earlier
				newCheckIn = checkIn.AddDate(0, 0, -1)
				newCheckOut = checkOut
			case 3:
				// Arrive 1 day later
				newCheckIn = checkIn.AddDate(0, 0, 1)
				newCheckOut = checkOut
			case 4:
				// Completely change dates (shift by 1 week)
				newCheckIn = checkIn.AddDate(0, 0, 7)
				newCheckOut = checkOut.AddDate(0, 0, 7)
			}

			// Small random delay to simulate human variability
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)

			// Try to update the booking
			err := client.UpdateBooking(booking.BookingID, newCheckIn, newCheckOut)

			if err == nil {
				log.Printf("Client %d successfully modified booking %d",
					clientID, booking.BookingID)

				// If modification succeeded, maybe try to get the updated booking
				if rand.Float32() < 0.7 {
					updatedBooking, _ := client.GetBooking(booking.BookingID)
					if updatedBooking != nil {
						log.Printf("Client %d retrieved modified booking: %s to %s",
							clientID, updatedBooking.CheckInDate.Format("2006-01-02"),
							updatedBooking.CheckOutDate.Format("2006-01-02"))
					}
				}
			} else {
				log.Printf("Client %d failed to modify booking: %v", clientID, err)

				// If modification failed, 50% chance to retry with slightly different dates
				if rand.Float32() < 0.5 {
					// Adjust dates slightly
					retryCheckIn := newCheckIn.AddDate(0, 0, 1)
					retryCheckOut := newCheckOut.AddDate(0, 0, 1)

					// Wait a bit before retry
					time.Sleep(300 * time.Millisecond)

					err = client.UpdateBooking(booking.BookingID, retryCheckIn, retryCheckOut)
					if err == nil {
						log.Printf("Client %d successful on retry", clientID)
					}
				}
			}
		}(i)
	}

	wg.Wait()

	// Now try to cancel the booking while simultaneously trying to update it
	var cancelWg sync.WaitGroup
	cancelWg.Add(2)

	// Thread trying to cancel
	go func() {
		defer cancelWg.Done()

		client := NewClient(100, baseURL, stats)
		err := client.CancelBooking(booking.BookingID)

		if err == nil {
			log.Printf("Successfully cancelled booking %d", booking.BookingID)

			// Try to process refund
			client.ProcessRefund(booking.BookingID)
		} else {
			log.Printf("Failed to cancel booking: %v", err)
		}
	}()

	// Thread trying to update at the same time
	go func() {
		defer cancelWg.Done()

		client := NewClient(101, baseURL, stats)

		// Completely different dates
		newCheckIn := checkIn.AddDate(0, 0, 14)    // 2 weeks later
		newCheckOut := newCheckIn.AddDate(0, 0, 3) // 3-day stay

		err := client.UpdateBooking(booking.BookingID, newCheckIn, newCheckOut)

		if err == nil {
			log.Printf("Successfully updated booking %d while cancellation in progress",
				booking.BookingID)
		} else {
			log.Printf("Failed to update booking (expected if cancelled): %v", err)
		}
	}()

	cancelWg.Wait()
	log.Printf("Booking modification conflicts scenario completed")
}

// Run a mix of clients with different behaviors
func runMixedClientScenario(baseURL string, numClients int, duration time.Duration) {
	stats := NewTestStatistics()

	log.Printf("Starting mixed client scenario with %d clients for %v",
		numClients, duration)

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(numClients)

	// Start clients with different behavior patterns
	for i := 1; i <= numClients; i++ {
		client := NewClient(i, baseURL, stats)
		go client.Run(ctx, &wg)
	}

	// Create a ticker to periodically report statistics
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				printRuntimeStatistics(stats)
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()

	// Wait for completion or timeout
	wg.Wait()
	log.Printf("Mixed client scenario completed after %v", duration)

	// Final statistics
	printScenarioStatistics("mixed_clients", stats)
}

func generateTimelineReport(stats *TestStatistics, scenarioName string) {
	// Finish the test tracking
	stats.TimelineGenerator.FinishTest()

	// Set the template path based on your project structure
	stats.TimelineGenerator.SetTemplatePath("web/templates/reports/timeline.html")

	// Create output directory if it doesn't exist
	outputDir := "test-results"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Printf("Failed to create output directory: %v", err)
		return
	}

	// Generate report with timestamp
	timestamp := time.Now().Format("20060102-150405")
	outputPath := filepath.Join(outputDir, fmt.Sprintf("timeline_%s_%s.html", scenarioName, timestamp))

	if err := stats.TimelineGenerator.GenerateReport(outputPath); err != nil {
		log.Printf("Failed to generate timeline report: %v", err)
		return
	}

	log.Printf("Timeline report generated: %s", outputPath)
}
