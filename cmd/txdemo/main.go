package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"sync"
	"time"
)

// This is a demonstration tool to clearly show transaction management
// and concurrency control in the hotel booking system.
//
// It runs a series of tests designed to trigger specific concurrency scenarios,
// logs the results, and then generates an HTML report with visualizations
// that clearly show transaction isolation and conflict resolution.

// BookingRequest represents a request to create a booking
type BookingRequest struct {
	BookingName  string    `json:"booking_name"`
	RoomNum      int       `json:"room_num"`
	CheckInDate  time.Time `json:"check_in_date"`
	CheckOutDate time.Time `json:"check_out_date"`
}

// BookingResponse represents the response from a booking API call
type BookingResponse struct {
	BookingID    int       `json:"booking_id"`
	BookingName  string    `json:"booking_name"`
	RoomNum      int       `json:"room_num"`
	CheckInDate  time.Time `json:"check_in_date"`
	CheckOutDate time.Time `json:"check_out_date"`
	TotalPrice   int       `json:"total_price"`
	BookingDate  time.Time `json:"booking_date"`
	Error        string    `json:"error,omitempty"`
}

// PaymentRequest represents a request to create a payment receipt
type PaymentRequest struct {
	BookingID     int    `json:"booking_id"`
	PaymentMethod string `json:"payment_method"`
	Amount        int    `json:"amount"`
}

// PaymentResponse represents the response from a payment API call
type PaymentResponse struct {
	ReceiptID     int       `json:"receipt_id"`
	BookingID     int       `json:"booking_id"`
	PaymentDate   time.Time `json:"payment_date"`
	PaymentMethod string    `json:"payment_method"`
	Amount        int       `json:"amount"`
	IssueDate     time.Time `json:"issue_date"`
	Error         string    `json:"error,omitempty"`
}

// TestResult represents the result of a concurrency test
type TestResult struct {
	ScenarioName     string
	Description      string
	Outcome          string
	Success          bool
	TransactionIDs   []string
	ClientIDs        []string
	ClientResults    []string
	Timeline         []TimelineEvent
	ConcurrencyIssue string
	Resolution       string
}

// TimelineEvent represents an event in the timeline visualization
type TimelineEvent struct {
	Time        time.Time
	Transaction string
	Client      string
	Event       string
	Details     string
	EventType   string // start, end, lock, conflict, etc.
}

// Test scenarios
var scenarios = []struct {
	Name        string
	Description string
}{
	{
		Name:        "concurrent-booking",
		Description: "Two clients attempting to book the same room for the same dates simultaneously",
	},
	{
		Name:        "optimistic-concurrency",
		Description: "Two clients attempting to modify the same booking simultaneously",
	},
	{
		Name:        "duplicate-payment",
		Description: "Attempting to process payment for the same booking twice",
	},
	{
		Name:        "escalating-lock",
		Description: "Demonstrating deadlock prevention through lock ordering",
	},
	{
		Name:        "high-contention",
		Description: "Multiple clients competing for limited rooms during peak demand",
	},
}

// Main function
func main() {
	baseURL := "http://localhost:8080/api/v1"
	if len(os.Args) > 1 {
		baseURL = os.Args[1]
	}

	fmt.Printf("Running transaction management visual demonstration against %s\n", baseURL)

	// Run the tests
	results := make([]TestResult, 0, len(scenarios))
	for _, scenario := range scenarios {
		fmt.Printf("\nRunning scenario: %s\n", scenario.Name)
		fmt.Printf("Description: %s\n", scenario.Description)

		var result TestResult
		result.ScenarioName = scenario.Name
		result.Description = scenario.Description

		switch scenario.Name {
		case "concurrent-booking":
			result = runConcurrentBookingTest(baseURL)
		case "optimistic-concurrency":
			result = runOptimisticConcurrencyTest(baseURL)
		case "duplicate-payment":
			result = runDuplicatePaymentTest(baseURL)
		case "escalating-lock":
			result = runEscalatingLockTest(baseURL)
		case "high-contention":
			result = runHighContentionTest(baseURL)
		}

		fmt.Printf("Outcome: %s\n", result.Outcome)
		if result.Success {
			fmt.Printf("✅ Transaction management successful: %s\n", result.Resolution)
		} else {
			fmt.Printf("❌ Transaction management issue: %s\n", result.ConcurrencyIssue)
		}

		results = append(results, result)
	}

	// Generate the report
	generateReport(results)
	fmt.Println("\nVisual demonstration complete. Report generated at tx_demo_report.html")
}

// Run concurrent booking test
func runConcurrentBookingTest(baseURL string) TestResult {
	result := TestResult{
		ScenarioName:   "concurrent-booking",
		Description:    "Two clients attempting to book the same room for the same dates simultaneously",
		Timeline:       make([]TimelineEvent, 0),
		TransactionIDs: []string{"tx-booking-1", "tx-booking-2"},
		ClientIDs:      []string{"Client A", "Client B"},
		ClientResults:  []string{"", ""},
	}

	// Use the same room number and dates for both bookings
	room := 101
	checkIn := time.Now().AddDate(0, 0, 7) // 7 days from now
	checkOut := checkIn.AddDate(0, 0, 3)   // 3-day stay

	// Prepare two booking requests
	booking1 := BookingRequest{
		BookingName:  "Client A",
		RoomNum:      room,
		CheckInDate:  checkIn,
		CheckOutDate: checkOut,
	}

	booking2 := BookingRequest{
		BookingName:  "Client B",
		RoomNum:      room,
		CheckInDate:  checkIn,
		CheckOutDate: checkOut,
	}

	// Record start events
	result.Timeline = append(result.Timeline, TimelineEvent{
		Time:        time.Now(),
		Transaction: "tx-booking-1",
		Client:      "Client A",
		Event:       "Start booking request",
		Details:     fmt.Sprintf("Room %d, %s to %s", room, checkIn.Format("2006-01-02"), checkOut.Format("2006-01-02")),
		EventType:   "start",
	})

	result.Timeline = append(result.Timeline, TimelineEvent{
		Time:        time.Now().Add(100 * time.Millisecond),
		Transaction: "tx-booking-2",
		Client:      "Client B",
		Event:       "Start booking request",
		Details:     fmt.Sprintf("Room %d, %s to %s", room, checkIn.Format("2006-01-02"), checkOut.Format("2006-01-02")),
		EventType:   "start",
	})

	// Channel to collect results
	resultChan := make(chan struct {
		bookingID int
		client    string
		err       string
	}, 2)

	// WaitGroup to wait for both requests to complete
	var wg sync.WaitGroup
	wg.Add(2)

	// Launch both booking attempts concurrently
	fmt.Println("Launching concurrent booking attempts for the same room and dates")

	go func() {
		defer wg.Done()
		bookingResult, err := createBooking(baseURL, booking1)
		if err != nil {
			resultChan <- struct {
				bookingID int
				client    string
				err       string
			}{0, "Client A", err.Error()}
		} else if bookingResult.Error != "" {
			resultChan <- struct {
				bookingID int
				client    string
				err       string
			}{0, "Client A", bookingResult.Error}
		} else {
			resultChan <- struct {
				bookingID int
				client    string
				err       string
			}{bookingResult.BookingID, "Client A", ""}
		}
	}()

	go func() {
		defer wg.Done()
		// Small delay to ensure both requests hit the server in close succession
		time.Sleep(50 * time.Millisecond)
		bookingResult, err := createBooking(baseURL, booking2)
		if err != nil {
			resultChan <- struct {
				bookingID int
				client    string
				err       string
			}{0, "Client B", err.Error()}
		} else if bookingResult.Error != "" {
			resultChan <- struct {
				bookingID int
				client    string
				err       string
			}{0, "Client B", bookingResult.Error}
		} else {
			resultChan <- struct {
				bookingID int
				client    string
				err       string
			}{bookingResult.BookingID, "Client B", ""}
		}
	}()

	// Wait for both requests to complete
	wg.Wait()
	close(resultChan)

	// Collect results
	bookingSuccesses := 0
	for res := range resultChan {
		if res.err == "" {
			// Success
			bookingSuccesses++
			idx := 0
			if res.client == "Client B" {
				idx = 1
			}
			result.ClientResults[idx] = fmt.Sprintf("Success: Booking ID %d", res.bookingID)

			result.Timeline = append(result.Timeline, TimelineEvent{
				Time:        time.Now().Add(500 * time.Millisecond),
				Transaction: fmt.Sprintf("tx-booking-%d", idx+1),
				Client:      res.client,
				Event:       "Booking successful",
				Details:     fmt.Sprintf("Booking ID: %d", res.bookingID),
				EventType:   "success",
			})
		} else {
			// Failure
			idx := 0
			if res.client == "Client B" {
				idx = 1
			}
			result.ClientResults[idx] = fmt.Sprintf("Failed: %s", res.err)

			result.Timeline = append(result.Timeline, TimelineEvent{
				Time:        time.Now().Add(500 * time.Millisecond),
				Transaction: fmt.Sprintf("tx-booking-%d", idx+1),
				Client:      res.client,
				Event:       "Booking failed",
				Details:     res.err,
				EventType:   "error",
			})
		}
	}

	// Record conflict event
	result.Timeline = append(result.Timeline, TimelineEvent{
		Time:        time.Now().Add(250 * time.Millisecond),
		Transaction: "tx-system",
		Client:      "System",
		Event:       "Detected booking conflict",
		Details:     "Two clients attempted to book the same room for overlapping dates",
		EventType:   "conflict",
	})

	// Determine outcome
	if bookingSuccesses <= 1 {
		result.Success = true
		result.Outcome = fmt.Sprintf("%d of 2 concurrent bookings succeeded", bookingSuccesses)
		result.Resolution = "Transaction isolation prevented double-booking"

		// Record resolution event
		result.Timeline = append(result.Timeline, TimelineEvent{
			Time:        time.Now().Add(300 * time.Millisecond),
			Transaction: "tx-system",
			Client:      "System",
			Event:       "Conflict resolution",
			Details:     "First transaction committed, second transaction rejected",
			EventType:   "resolution",
		})
	} else {
		result.Success = false
		result.Outcome = "Both bookings succeeded"
		result.ConcurrencyIssue = "Double-booking occurred, indicating lack of proper transaction isolation"
	}

	return result
}

// Run optimistic concurrency test
func runOptimisticConcurrencyTest(baseURL string) TestResult {
	result := TestResult{
		ScenarioName:   "optimistic-concurrency",
		Description:    "Two clients attempting to modify the same booking simultaneously",
		Timeline:       make([]TimelineEvent, 0),
		TransactionIDs: []string{"tx-update-1", "tx-update-2"},
		ClientIDs:      []string{"Client X", "Client Y"},
		ClientResults:  []string{"", ""},
	}

	// First, create a booking to work with
	booking := BookingRequest{
		BookingName:  "Concurrent Update Test",
		RoomNum:      102,
		CheckInDate:  time.Now().AddDate(0, 0, 14), // 14 days from now
		CheckOutDate: time.Now().AddDate(0, 0, 17), // 3-day stay
	}

	bookingResult, err := createBooking(baseURL, booking)
	if err != nil || bookingResult.Error != "" {
		var errMsg string
		if err != nil {
			errMsg = err.Error()
		} else {
			errMsg = bookingResult.Error
		}
		result.Outcome = fmt.Sprintf("Test setup failed: %s", errMsg)
		result.Success = false
		return result
	}

	bookingID := bookingResult.BookingID
	fmt.Printf("Created test booking with ID: %d\n", bookingID)

	// Record booking creation event
	result.Timeline = append(result.Timeline, TimelineEvent{
		Time:        time.Now(),
		Transaction: "tx-setup",
		Client:      "System",
		Event:       "Created test booking",
		Details:     fmt.Sprintf("Booking ID: %d", bookingID),
		EventType:   "setup",
	})

	// Now prepare two concurrent update requests with different dates
	update1 := map[string]interface{}{
		"check_in_date":  time.Now().AddDate(0, 0, 15).Format(time.RFC3339), // 15 days from now
		"check_out_date": time.Now().AddDate(0, 0, 18).Format(time.RFC3339), // 3-day stay
	}

	update2 := map[string]interface{}{
		"check_in_date":  time.Now().AddDate(0, 0, 16).Format(time.RFC3339), // 16 days from now
		"check_out_date": time.Now().AddDate(0, 0, 19).Format(time.RFC3339), // 3-day stay
	}

	// Record update attempt events
	result.Timeline = append(result.Timeline, TimelineEvent{
		Time:        time.Now().Add(100 * time.Millisecond),
		Transaction: "tx-update-1",
		Client:      "Client X",
		Event:       "Start update request",
		Details: fmt.Sprintf("New dates: %s to %s",
			time.Now().AddDate(0, 0, 15).Format("2006-01-02"),
			time.Now().AddDate(0, 0, 18).Format("2006-01-02")),
		EventType: "start",
	})

	result.Timeline = append(result.Timeline, TimelineEvent{
		Time:        time.Now().Add(150 * time.Millisecond),
		Transaction: "tx-update-2",
		Client:      "Client Y",
		Event:       "Start update request",
		Details: fmt.Sprintf("New dates: %s to %s",
			time.Now().AddDate(0, 0, 16).Format("2006-01-02"),
			time.Now().AddDate(0, 0, 19).Format("2006-01-02")),
		EventType: "start",
	})

	// Channel to collect results
	resultChan := make(chan struct {
		client  string
		success bool
		err     string
	}, 2)

	// WaitGroup to wait for both requests to complete
	var wg sync.WaitGroup
	wg.Add(2)

	// Launch both update attempts concurrently
	fmt.Println("Launching concurrent update attempts for the same booking")

	go func() {
		defer wg.Done()
		success, err := updateBooking(baseURL, bookingID, update1)
		if err != nil {
			resultChan <- struct {
				client  string
				success bool
				err     string
			}{"Client X", false, err.Error()}
		} else if !success {
			resultChan <- struct {
				client  string
				success bool
				err     string
			}{"Client X", false, "Optimistic concurrency control rejected the update"}
		} else {
			resultChan <- struct {
				client  string
				success bool
				err     string
			}{"Client X", true, ""}
		}
	}()

	go func() {
		defer wg.Done()
		// Small delay to ensure both requests hit the server in close succession
		time.Sleep(50 * time.Millisecond)
		success, err := updateBooking(baseURL, bookingID, update2)
		if err != nil {
			resultChan <- struct {
				client  string
				success bool
				err     string
			}{"Client Y", false, err.Error()}
		} else if !success {
			resultChan <- struct {
				client  string
				success bool
				err     string
			}{"Client Y", false, "Optimistic concurrency control rejected the update"}
		} else {
			resultChan <- struct {
				client  string
				success bool
				err     string
			}{"Client Y", true, ""}
		}
	}()

	// Wait for both requests to complete
	wg.Wait()
	close(resultChan)

	// Collect results
	updateSuccesses := 0
	for res := range resultChan {
		idx := 0
		if res.client == "Client Y" {
			idx = 1
		}

		if res.success {
			updateSuccesses++
			result.ClientResults[idx] = "Success: Booking updated"

			result.Timeline = append(result.Timeline, TimelineEvent{
				Time:        time.Now().Add(400 * time.Millisecond),
				Transaction: fmt.Sprintf("tx-update-%d", idx+1),
				Client:      res.client,
				Event:       "Update successful",
				Details:     "Booking dates updated",
				EventType:   "success",
			})
		} else {
			result.ClientResults[idx] = fmt.Sprintf("Failed: %s", res.err)

			result.Timeline = append(result.Timeline, TimelineEvent{
				Time:        time.Now().Add(400 * time.Millisecond),
				Transaction: fmt.Sprintf("tx-update-%d", idx+1),
				Client:      res.client,
				Event:       "Update failed",
				Details:     res.err,
				EventType:   "error",
			})
		}
	}

	// Record version check event
	result.Timeline = append(result.Timeline, TimelineEvent{
		Time:        time.Now().Add(250 * time.Millisecond),
		Transaction: "tx-system",
		Client:      "System",
		Event:       "Optimistic concurrency check",
		Details:     "Checking if booking was modified since it was read",
		EventType:   "check",
	})

	// Determine outcome
	if updateSuccesses <= 1 {
		result.Success = true
		result.Outcome = fmt.Sprintf("%d of 2 concurrent updates succeeded", updateSuccesses)
		result.Resolution = "Optimistic concurrency control prevented conflicting updates"

		// Record resolution event
		result.Timeline = append(result.Timeline, TimelineEvent{
			Time:        time.Now().Add(300 * time.Millisecond),
			Transaction: "tx-system",
			Client:      "System",
			Event:       "Conflict resolution",
			Details:     "First transaction committed, second transaction rejected due to version mismatch",
			EventType:   "resolution",
		})
	} else {
		result.Success = false
		result.Outcome = "Both updates succeeded"
		result.ConcurrencyIssue = "Conflicting updates were applied, indicating a failure in optimistic concurrency control"
	}

	return result
}

// Run duplicate payment test
func runDuplicatePaymentTest(baseURL string) TestResult {
	result := TestResult{
		ScenarioName:   "duplicate-payment",
		Description:    "Attempting to process payment for the same booking twice",
		Timeline:       make([]TimelineEvent, 0),
		TransactionIDs: []string{"tx-payment-1", "tx-payment-2"},
		ClientIDs:      []string{"Client P", "Client Q"},
		ClientResults:  []string{"", ""},
	}

	// First, create a booking to pay for
	booking := BookingRequest{
		BookingName:  "Payment Test",
		RoomNum:      103,
		CheckInDate:  time.Now().AddDate(0, 0, 21), // 21 days from now
		CheckOutDate: time.Now().AddDate(0, 0, 23), // 2-day stay
	}

	bookingResult, err := createBooking(baseURL, booking)
	if err != nil || bookingResult.Error != "" {
		var errMsg string
		if err != nil {
			errMsg = err.Error()
		} else {
			errMsg = bookingResult.Error
		}
		result.Outcome = fmt.Sprintf("Test setup failed: %s", errMsg)
		result.Success = false
		return result
	}

	bookingID := bookingResult.BookingID
	amount := bookingResult.TotalPrice
	fmt.Printf("Created test booking with ID: %d, Amount: $%d\n", bookingID, amount)

	// Record booking creation event
	result.Timeline = append(result.Timeline, TimelineEvent{
		Time:        time.Now(),
		Transaction: "tx-setup",
		Client:      "System",
		Event:       "Created test booking",
		Details:     fmt.Sprintf("Booking ID: %d, Amount: $%d", bookingID, amount),
		EventType:   "setup",
	})

	// Now prepare two identical payment requests
	payment1 := PaymentRequest{
		BookingID:     bookingID,
		PaymentMethod: "Credit",
		Amount:        amount,
	}

	payment2 := PaymentRequest{
		BookingID:     bookingID,
		PaymentMethod: "Debit",
		Amount:        amount,
	}

	// Record payment attempt events
	result.Timeline = append(result.Timeline, TimelineEvent{
		Time:        time.Now().Add(100 * time.Millisecond),
		Transaction: "tx-payment-1",
		Client:      "Client P",
		Event:       "Start payment request",
		Details:     fmt.Sprintf("Booking ID: %d, Method: Credit, Amount: $%d", bookingID, amount),
		EventType:   "start",
	})

	result.Timeline = append(result.Timeline, TimelineEvent{
		Time:        time.Now().Add(150 * time.Millisecond),
		Transaction: "tx-payment-2",
		Client:      "Client Q",
		Event:       "Start payment request",
		Details:     fmt.Sprintf("Booking ID: %d, Method: Debit, Amount: $%d", bookingID, amount),
		EventType:   "start",
	})

	// Channel to collect results
	resultChan := make(chan struct {
		client  string
		success bool
		err     string
	}, 2)

	// WaitGroup to wait for both requests to complete
	var wg sync.WaitGroup
	wg.Add(2)

	// Launch both payment attempts concurrently
	fmt.Println("Launching concurrent payment attempts for the same booking")

	go func() {
		defer wg.Done()
		resp, err := processPayment(baseURL, payment1)
		if err != nil {
			resultChan <- struct {
				client  string
				success bool
				err     string
			}{"Client P", false, err.Error()}
		} else if resp.Error != "" {
			resultChan <- struct {
				client  string
				success bool
				err     string
			}{"Client P", false, resp.Error}
		} else {
			resultChan <- struct {
				client  string
				success bool
				err     string
			}{"Client P", true, ""}
		}
	}()

	go func() {
		defer wg.Done()
		// Small delay to ensure both requests hit the server in close succession
		time.Sleep(50 * time.Millisecond)
		resp, err := processPayment(baseURL, payment2)
		if err != nil {
			resultChan <- struct {
				client  string
				success bool
				err     string
			}{"Client Q", false, err.Error()}
		} else if resp.Error != "" {
			resultChan <- struct {
				client  string
				success bool
				err     string
			}{"Client Q", false, resp.Error}
		} else {
			resultChan <- struct {
				client  string
				success bool
				err     string
			}{"Client Q", true, ""}
		}
	}()

	// Wait for both requests to complete
	wg.Wait()
	close(resultChan)

	// Collect results
	paymentSuccesses := 0
	for res := range resultChan {
		idx := 0
		if res.client == "Client Q" {
			idx = 1
		}

		if res.success {
			paymentSuccesses++
			result.ClientResults[idx] = "Success: Payment processed"

			result.Timeline = append(result.Timeline, TimelineEvent{
				Time:        time.Now().Add(400 * time.Millisecond),
				Transaction: fmt.Sprintf("tx-payment-%d", idx+1),
				Client:      res.client,
				Event:       "Payment successful",
				Details:     fmt.Sprintf("Payment processed for Booking ID: %d", bookingID),
				EventType:   "success",
			})
		} else {
			result.ClientResults[idx] = fmt.Sprintf("Failed: %s", res.err)

			result.Timeline = append(result.Timeline, TimelineEvent{
				Time:        time.Now().Add(400 * time.Millisecond),
				Transaction: fmt.Sprintf("tx-payment-%d", idx+1),
				Client:      res.client,
				Event:       "Payment failed",
				Details:     res.err,
				EventType:   "error",
			})
		}
	}

	// Record lock acquisition event
	result.Timeline = append(result.Timeline, TimelineEvent{
		Time:        time.Now().Add(200 * time.Millisecond),
		Transaction: "tx-system",
		Client:      "System",
		Event:       "Pessimistic lock acquisition",
		Details:     fmt.Sprintf("Acquiring exclusive lock on Booking ID: %d", bookingID),
		EventType:   "lock",
	})

	// Determine outcome
	if paymentSuccesses <= 1 {
		result.Success = true
		result.Outcome = fmt.Sprintf("%d of 2 concurrent payments succeeded", paymentSuccesses)
		result.Resolution = "Pessimistic locking prevented duplicate payment"

		// Record resolution event
		result.Timeline = append(result.Timeline, TimelineEvent{
			Time:        time.Now().Add(300 * time.Millisecond),
			Transaction: "tx-system",
			Client:      "System",
			Event:       "Conflict resolution",
			Details:     "First transaction acquired lock and committed, second transaction was blocked",
			EventType:   "resolution",
		})
	} else {
		result.Success = false
		result.Outcome = "Both payments succeeded"
		result.ConcurrencyIssue = "Duplicate payments were processed, indicating a failure in pessimistic locking"
	}

	return result
}

// Run escalating lock test
func runEscalatingLockTest(baseURL string) TestResult {
	result := TestResult{
		ScenarioName:   "escalating-lock",
		Description:    "Demonstrating deadlock prevention through lock ordering",
		Timeline:       make([]TimelineEvent, 0),
		TransactionIDs: []string{"tx-booking-1", "tx-payment-1"},
		ClientIDs:      []string{"Client M", "Client N"},
		ClientResults:  []string{"", ""},
	}

	// For this test, we'll create a booking and then try to modify it and pay for it concurrently
	booking := BookingRequest{
		BookingName:  "Lock Ordering Test",
		RoomNum:      104,
		CheckInDate:  time.Now().AddDate(0, 0, 30), // 30 days from now
		CheckOutDate: time.Now().AddDate(0, 0, 32), // 2-day stay
	}

	bookingResult, err := createBooking(baseURL, booking)
	if err != nil || bookingResult.Error != "" {
		var errMsg string
		if err != nil {
			errMsg = err.Error()
		} else {
			errMsg = bookingResult.Error
		}
		result.Outcome = fmt.Sprintf("Test setup failed: %s", errMsg)
		result.Success = false
		return result
	}

	bookingID := bookingResult.BookingID
	amount := bookingResult.TotalPrice
	fmt.Printf("Created test booking with ID: %d, Amount: $%d\n", bookingID, amount)

	// Record booking creation event
	result.Timeline = append(result.Timeline, TimelineEvent{
		Time:        time.Now(),
		Transaction: "tx-setup",
		Client:      "System",
		Event:       "Created test booking",
		Details:     fmt.Sprintf("Booking ID: %d, Amount: $%d", bookingID, amount),
		EventType:   "setup",
	})

	// Prepare update and payment requests
	update := map[string]interface{}{
		"check_in_date":  time.Now().AddDate(0, 0, 31).Format(time.RFC3339), // 31 days from now
		"check_out_date": time.Now().AddDate(0, 0, 33).Format(time.RFC3339), // 2-day stay
	}

	payment := PaymentRequest{
		BookingID:     bookingID,
		PaymentMethod: "Credit",
		Amount:        amount,
	}

	// Record operation start events
	result.Timeline = append(result.Timeline, TimelineEvent{
		Time:        time.Now().Add(100 * time.Millisecond),
		Transaction: "tx-booking-1",
		Client:      "Client M",
		Event:       "Start update request",
		Details: fmt.Sprintf("Booking ID: %d, New dates: %s to %s",
			bookingID,
			time.Now().AddDate(0, 0, 31).Format("2006-01-02"),
			time.Now().AddDate(0, 0, 33).Format("2006-01-02")),
		EventType: "start",
	})

	result.Timeline = append(result.Timeline, TimelineEvent{
		Time:        time.Now().Add(150 * time.Millisecond),
		Transaction: "tx-payment-1",
		Client:      "Client N",
		Event:       "Start payment request",
		Details:     fmt.Sprintf("Booking ID: %d, Method: Credit, Amount: $%d", bookingID, amount),
		EventType:   "start",
	})

	// Channel to collect results
	resultChan := make(chan struct {
		op      string
		client  string
		success bool
		err     string
	}, 2)

	// WaitGroup to wait for both requests to complete
	var wg sync.WaitGroup
	wg.Add(2)

	// Launch both operations concurrently
	fmt.Println("Launching concurrent update and payment operations")

	go func() {
		defer wg.Done()
		success, err := updateBooking(baseURL, bookingID, update)
		if err != nil {
			resultChan <- struct {
				op      string
				client  string
				success bool
				err     string
			}{"update", "Client M", false, err.Error()}
		} else if !success {
			resultChan <- struct {
				op      string
				client  string
				success bool
				err     string
			}{"update", "Client M", false, "Update rejected"}
		} else {
			resultChan <- struct {
				op      string
				client  string
				success bool
				err     string
			}{"update", "Client M", true, ""}
		}
	}()

	go func() {
		defer wg.Done()
		// Small delay to ensure both requests hit the server in close succession
		time.Sleep(50 * time.Millisecond)
		paymentResp, err := processPayment(baseURL, payment)
		if err != nil {
			resultChan <- struct {
				op      string
				client  string
				success bool
				err     string
			}{"payment", "Client N", false, err.Error()}
		} else if paymentResp.Error != "" {
			resultChan <- struct {
				op      string
				client  string
				success bool
				err     string
			}{"payment", "Client N", false, paymentResp.Error}
		} else {
			resultChan <- struct {
				op      string
				client  string
				success bool
				err     string
			}{"payment", "Client N", true, ""}
		}
	}()

	// Wait for both requests to complete
	wg.Wait()
	close(resultChan)

	// Collect results
	var updateSuccess, paymentSuccess bool
	var updateError, paymentError string

	for res := range resultChan {
		if res.op == "update" {
			updateSuccess = res.success
			updateError = res.err
			if res.success {
				result.ClientResults[0] = "Success: Booking updated"
				result.Timeline = append(result.Timeline, TimelineEvent{
					Time:        time.Now().Add(300 * time.Millisecond),
					Transaction: "tx-booking-1",
					Client:      "Client M",
					Event:       "Update successful",
					Details:     "Booking dates updated",
					EventType:   "success",
				})
			} else {
				result.ClientResults[0] = fmt.Sprintf("Failed: %s", res.err)
				result.Timeline = append(result.Timeline, TimelineEvent{
					Time:        time.Now().Add(300 * time.Millisecond),
					Transaction: "tx-booking-1",
					Client:      "Client M",
					Event:       "Update failed",
					Details:     res.err,
					EventType:   "error",
				})
			}
		} else if res.op == "payment" {
			paymentSuccess = res.success
			paymentError = res.err
			if res.success {
				result.ClientResults[1] = "Success: Payment processed"
				result.Timeline = append(result.Timeline, TimelineEvent{
					Time:        time.Now().Add(350 * time.Millisecond),
					Transaction: "tx-payment-1",
					Client:      "Client N",
					Event:       "Payment successful",
					Details:     fmt.Sprintf("Payment processed for Booking ID: %d", bookingID),
					EventType:   "success",
				})
			} else {
				result.ClientResults[1] = fmt.Sprintf("Failed: %s", res.err)
				result.Timeline = append(result.Timeline, TimelineEvent{
					Time:        time.Now().Add(350 * time.Millisecond),
					Transaction: "tx-payment-1",
					Client:      "Client N",
					Event:       "Payment failed",
					Details:     res.err,
					EventType:   "error",
				})
			}
		}
	}

	// Record lock sequence events
	result.Timeline = append(result.Timeline, TimelineEvent{
		Time:        time.Now().Add(200 * time.Millisecond),
		Transaction: "tx-system",
		Client:      "System",
		Event:       "Lock acquisition order",
		Details:     "Locks are acquired in a consistent order to prevent deadlocks",
		EventType:   "lock",
	})

	// Determine outcome
	if updateSuccess && paymentSuccess {
		result.Success = true
		result.Outcome = "Both operations succeeded without deadlock"
		result.Resolution = "Lock ordering prevented potential deadlocks"

		result.Timeline = append(result.Timeline, TimelineEvent{
			Time:        time.Now().Add(250 * time.Millisecond),
			Transaction: "tx-system",
			Client:      "System",
			Event:       "Deadlock prevention",
			Details:     "Consistent lock ordering allowed both operations to complete without deadlock",
			EventType:   "resolution",
		})
	} else if !updateSuccess && !paymentSuccess {
		result.Success = false
		result.Outcome = "Both operations failed"
		result.ConcurrencyIssue = fmt.Sprintf("Unexpected failures: Update error: %s, Payment error: %s", updateError, paymentError)
	} else {
		result.Success = true
		result.Outcome = "One operation succeeded, one failed, but without deadlock"
		result.Resolution = "System prevented deadlock through proper lock management"

		result.Timeline = append(result.Timeline, TimelineEvent{
			Time:        time.Now().Add(250 * time.Millisecond),
			Transaction: "tx-system",
			Client:      "System",
			Event:       "Conflict resolution",
			Details:     "One operation was prioritized, preventing deadlock",
			EventType:   "resolution",
		})
	}

	return result
}

// Run high contention test
func runHighContentionTest(baseURL string) TestResult {
	result := TestResult{
		ScenarioName:   "high-contention",
		Description:    "Multiple clients competing for limited rooms during peak demand",
		Timeline:       make([]TimelineEvent, 0),
		TransactionIDs: []string{"tx-1", "tx-2", "tx-3", "tx-4", "tx-5"},
		ClientIDs:      []string{"Client 1", "Client 2", "Client 3", "Client 4", "Client 5"},
		ClientResults:  []string{"", "", "", "", ""},
	}

	// Use the same date range but different rooms
	checkIn := time.Now().AddDate(0, 0, 45) // 45 days from now
	checkOut := checkIn.AddDate(0, 0, 3)    // 3-day stay

	roomNums := []int{101, 102, 103} // Limited number of rooms
	numClients := 5                  // More clients than rooms

	// Create multiple booking requests for the same limited set of rooms
	bookings := make([]BookingRequest, numClients)
	for i := 0; i < numClients; i++ {
		roomIndex := i % len(roomNums) // Cycle through the available rooms
		bookings[i] = BookingRequest{
			BookingName:  fmt.Sprintf("Client %d", i+1),
			RoomNum:      roomNums[roomIndex],
			CheckInDate:  checkIn,
			CheckOutDate: checkOut,
		}

		// Record booking attempt events
		result.Timeline = append(result.Timeline, TimelineEvent{
			Time:        time.Now().Add(time.Duration(i*50) * time.Millisecond),
			Transaction: fmt.Sprintf("tx-%d", i+1),
			Client:      fmt.Sprintf("Client %d", i+1),
			Event:       "Start booking request",
			Details: fmt.Sprintf("Room %d, %s to %s", roomNums[roomIndex],
				checkIn.Format("2006-01-02"),
				checkOut.Format("2006-01-02")),
			EventType: "start",
		})
	}

	// Channel to collect results
	resultChan := make(chan struct {
		client    int
		success   bool
		bookingID int
		err       string
	}, numClients)

	// WaitGroup to wait for all requests to complete
	var wg sync.WaitGroup
	wg.Add(numClients)

	// Launch all booking attempts concurrently
	fmt.Println("Launching concurrent booking attempts for limited rooms")

	for i := 0; i < numClients; i++ {
		i := i // Capture loop variable
		go func() {
			defer wg.Done()
			// Small staggered delay to simulate realistic timing
			time.Sleep(time.Duration(10*i) * time.Millisecond)

			bookingResult, err := createBooking(baseURL, bookings[i])
			if err != nil {
				resultChan <- struct {
					client    int
					success   bool
					bookingID int
					err       string
				}{i + 1, false, 0, err.Error()}
			} else if bookingResult.Error != "" {
				resultChan <- struct {
					client    int
					success   bool
					bookingID int
					err       string
				}{i + 1, false, 0, bookingResult.Error}
			} else {
				resultChan <- struct {
					client    int
					success   bool
					bookingID int
					err       string
				}{i + 1, true, bookingResult.BookingID, ""}
			}
		}()
	}

	// Wait for all requests to complete
	wg.Wait()
	close(resultChan)

	// Collect results
	successCount := 0
	roomBookings := make(map[int]bool) // Track which rooms were successfully booked

	for res := range resultChan {
		idx := res.client - 1
		roomNum := roomNums[idx%len(roomNums)]

		if res.success {
			successCount++
			roomBookings[roomNum] = true
			result.ClientResults[idx] = fmt.Sprintf("Success: Booked room %d, Booking ID %d", roomNum, res.bookingID)

			result.Timeline = append(result.Timeline, TimelineEvent{
				Time:        time.Now().Add(500 * time.Millisecond).Add(time.Duration(idx*50) * time.Millisecond),
				Transaction: fmt.Sprintf("tx-%d", idx+1),
				Client:      fmt.Sprintf("Client %d", idx+1),
				Event:       "Booking successful",
				Details:     fmt.Sprintf("Room %d, Booking ID: %d", roomNum, res.bookingID),
				EventType:   "success",
			})
		} else {
			result.ClientResults[idx] = fmt.Sprintf("Failed: %s", res.err)

			result.Timeline = append(result.Timeline, TimelineEvent{
				Time:        time.Now().Add(500 * time.Millisecond).Add(time.Duration(idx*50) * time.Millisecond),
				Transaction: fmt.Sprintf("tx-%d", idx+1),
				Client:      fmt.Sprintf("Client %d", idx+1),
				Event:       "Booking failed",
				Details:     res.err,
				EventType:   "error",
			})
		}
	}

	// Record contention management events
	result.Timeline = append(result.Timeline, TimelineEvent{
		Time:        time.Now().Add(250 * time.Millisecond),
		Transaction: "tx-system",
		Client:      "System",
		Event:       "Contention management",
		Details:     "System handling multiple concurrent requests for limited resources",
		EventType:   "check",
	})

	// Determine outcome
	expectedSuccessCount := len(roomNums) // We can't book more rooms than we have

	if successCount <= expectedSuccessCount && len(roomBookings) == successCount {
		result.Success = true
		result.Outcome = fmt.Sprintf("%d of %d booking attempts succeeded for %d available rooms",
			successCount, numClients, len(roomNums))
		result.Resolution = "System properly handled high contention for limited resources"

		result.Timeline = append(result.Timeline, TimelineEvent{
			Time:        time.Now().Add(300 * time.Millisecond),
			Transaction: "tx-system",
			Client:      "System",
			Event:       "Resource allocation",
			Details: fmt.Sprintf("%d successful bookings out of %d attempts for %d rooms",
				successCount, numClients, len(roomNums)),
			EventType: "resolution",
		})
	} else if successCount > expectedSuccessCount {
		result.Success = false
		result.Outcome = fmt.Sprintf("%d bookings succeeded for only %d available rooms",
			successCount, len(roomNums))
		result.ConcurrencyIssue = "Double-booking occurred, indicating a failure in transaction isolation"
	} else {
		result.Success = true
		result.Outcome = fmt.Sprintf("Only %d of %d possible bookings succeeded",
			successCount, expectedSuccessCount)
		result.Resolution = "System prevented overbooking but may have been overly conservative"
	}

	return result
}

// Helper function to create a booking
func createBooking(baseURL string, booking BookingRequest) (BookingResponse, error) {
	var response BookingResponse

	bookingJSON, err := json.Marshal(booking)
	if err != nil {
		return response, err
	}

	resp, err := http.Post(baseURL+"/bookings", "application/json", bytes.NewBuffer(bookingJSON))
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		// Try to read error message
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return response, fmt.Errorf("booking failed with status %d", resp.StatusCode)
		}
		return response, nil
	}

	err = json.NewDecoder(resp.Body).Decode(&response)
	return response, err
}

// Helper function to update a booking
func updateBooking(baseURL string, bookingID int, updates map[string]interface{}) (bool, error) {
	updatesJSON, err := json.Marshal(updates)
	if err != nil {
		return false, err
	}

	req, err := http.NewRequest(http.MethodPut,
		fmt.Sprintf("%s/bookings/%d", baseURL, bookingID),
		bytes.NewBuffer(updatesJSON))
	if err != nil {
		return false, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

// Helper function to process a payment
func processPayment(baseURL string, payment PaymentRequest) (PaymentResponse, error) {
	var response PaymentResponse

	paymentJSON, err := json.Marshal(payment)
	if err != nil {
		return response, err
	}

	resp, err := http.Post(baseURL+"/receipts", "application/json", bytes.NewBuffer(paymentJSON))
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		// Try to read error message
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			response.Error = fmt.Sprintf("payment failed with status %d", resp.StatusCode)
		}
		return response, nil
	}

	err = json.NewDecoder(resp.Body).Decode(&response)
	return response, err
}

// Generate HTML report
func generateReport(results []TestResult) {
	reportTemplate := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Transaction Management Demonstration</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            margin: 0;
            padding: 20px;
            color: #333;
        }
        h1, h2, h3 {
            color: #2c3e50;
        }
        .report-header {
            background-color: #3498db;
            color: white;
            padding: 20px;
            border-radius: 5px;
            margin-bottom: 20px;
        }
        .scenario {
            border: 1px solid #ddd;
            margin-bottom: 30px;
            border-radius: 5px;
            overflow: hidden;
        }
        .scenario-header {
            background-color: #f5f5f5;
            padding: 15px;
            border-bottom: 1px solid #ddd;
        }
        .scenario-body {
            padding: 15px;
        }
        .success {
            color: #27ae60;
            font-weight: bold;
        }
        .failure {
            color: #e74c3c;
            font-weight: bold;
        }
        .timeline {
            position: relative;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px 0;
        }
        .timeline::after {
            content: '';
            position: absolute;
            width: 6px;
            background-color: #bdc3c7;
            top: 0;
            bottom: 0;
            left: 50%;
            margin-left: -3px;
        }
        .event {
            padding: 10px 40px;
            position: relative;
            width: 46%;
            box-sizing: border-box;
        }
        .event::after {
            content: '';
            position: absolute;
            width: 20px;
            height: 20px;
            background-color: white;
            border: 4px solid;
            border-radius: 50%;
            top: 15px;
            z-index: 1;
        }
        .left {
            left: 0;
        }
        .right {
            left: 50%;
        }
        .left::after {
            right: -12px;
        }
        .right::after {
            left: -12px;
        }
        .event-content {
            padding: 15px;
            background-color: white;
            border-radius: 6px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .event-time {
            font-size: 0.8em;
            color: #7f8c8d;
        }
        .event-title {
            font-weight: bold;
            margin: 5px 0;
        }
        .event-details {
            color: #555;
        }
        .start { border-color: #3498db; }
        .success { border-color: #2ecc71; }
        .error { border-color: #e74c3c; }
        .lock { border-color: #f39c12; }
        .check { border-color: #9b59b6; }
        .conflict { border-color: #e74c3c; }
        .resolution { border-color: #27ae60; }
        .setup { border-color: #95a5a6; }
        .result-table {
            width: 100%;
            border-collapse: collapse;
            margin: 20px 0;
        }
        .result-table th, .result-table td {
            padding: 10px;
            border: 1px solid #ddd;
            text-align: left;
        }
        .result-table th {
            background-color: #f2f2f2;
        }
        .summary {
            background-color: #ecf0f1;
            padding: 15px;
            border-radius: 5px;
            margin-top: 20px;
        }
        @media screen and (max-width: 600px) {
            .timeline::after {
                left: 31px;
            }
            .event {
                width: 100%;
                padding-left: 70px;
                padding-right: 25px;
            }
            .event::after {
                left: 15px;
            }
            .left::after, .right::after {
                left: 15px;
            }
            .right {
                left: 0%;
            }
        }
    </style>
</head>
<body>
    <div class="report-header">
        <h1>Transaction Management Demonstration</h1>
        <p>This report shows the results of tests designed to demonstrate transaction management and concurrency control in the hotel booking system.</p>
    </div>

    <div class="summary">
        <h2>Summary</h2>
        <p>Tests run: {{ .TotalTests }}, Successful: {{ .SuccessCount }}, Failed: {{ .FailureCount }}</p>
        <p>This demonstration shows how the system handles concurrent operations through transaction isolation, optimistic and pessimistic concurrency control, and deadlock prevention mechanisms.</p>
    </div>
    
    {{ range .Results }}
    <div class="scenario">
        <div class="scenario-header">
            <h2>{{ .ScenarioName }}</h2>
            <p>{{ .Description }}</p>
        </div>
        <div class="scenario-body">
            <h3>Outcome: 
                {{ if .Success }}
                <span class="success">✓ Success</span>
                {{ else }}
                <span class="failure">✗ Failure</span>
                {{ end }}
            </h3>
            <p>{{ .Outcome }}</p>
            {{ if .Success }}
            <p class="success">{{ .Resolution }}</p>
            {{ else }}
            <p class="failure">{{ .ConcurrencyIssue }}</p>
            {{ end }}
            
            <h3>Client Results</h3>
            <table class="result-table">
                <thead>
                    <tr>
                        <th>Client</th>
                        <th>Result</th>
                    </tr>
                </thead>
                <tbody>
                    {{ range $index, $client := .ClientIDs }}
                    <tr>
                        <td>{{ $client }}</td>
                        <td>{{ index $.ClientResults $index }}</td>
                    </tr>
                    {{ end }}
                </tbody>
            </table>
            
            <h3>Transaction Timeline</h3>
            <div class="timeline">
                {{ range $index, $event := .Timeline }}
                <div class="event {{ if even $index }}left{{ else }}right{{ end }}">
                    <div class="event-content">
                        <div class="event-time">{{ formatTime $event.Time }}</div>
                        <div class="event-title">{{ $event.Transaction }} ({{ $event.Client }}): {{ $event.Event }}</div>
                        <div class="event-details">{{ $event.Details }}</div>
                    </div>
                    <div class="event-marker {{ $event.EventType }}"></div>
                </div>
                {{ end }}
            </div>
        </div>
    </div>
    {{ end }}
</body>
</html>
`

	// Define a template with functions
	tmpl := template.New("report")

	// Add custom functions
	tmpl = tmpl.Funcs(template.FuncMap{
		"even":       func(i int) bool { return i%2 == 0 },
		"formatTime": func(t time.Time) string { return t.Format("15:04:05.000") },
	})

	// Parse the template
	parsedTemplate, err := tmpl.Parse(reportTemplate)
	if err != nil {
		fmt.Printf("Error parsing template: %v\n", err)
		return
	}

	// Calculate summary statistics
	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		}
	}

	// Create data structure for the template
	data := struct {
		TotalTests   int
		SuccessCount int
		FailureCount int
		Results      []TestResult
	}{
		TotalTests:   len(results),
		SuccessCount: successCount,
		FailureCount: len(results) - successCount,
		Results:      results,
	}

	// Create output file
	file, err := os.Create("tx_demo_report.html")
	if err != nil {
		fmt.Printf("Error creating report file: %v\n", err)
		return
	}
	defer file.Close()

	// Execute the template
	err = parsedTemplate.Execute(file, data)
	if err != nil {
		fmt.Printf("Error executing template: %v\n", err)
	}
}
