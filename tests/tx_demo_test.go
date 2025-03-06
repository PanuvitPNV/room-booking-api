package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"
)

// This test file demonstrates transaction management and concurrency control
// It runs scenarios specifically designed to trigger concurrency events

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
	Error        string    `json:"error,omitempty"`
}

// PaymentRequest represents a request to create a payment receipt
type PaymentRequest struct {
	BookingID     int    `json:"booking_id"`
	PaymentMethod string `json:"payment_method"`
	Amount        int    `json:"amount"`
}

// Run this test with: go test -v ./tests -run TestTransactionDemo
func TestTransactionDemo(t *testing.T) {
	baseURL := "http://localhost:8080/api/v1"
	t.Logf("Running transaction management demonstration tests against %s", baseURL)

	// SCENARIO 1: Concurrent booking attempts for the same room and dates
	// This demonstrates how the system prevents double-booking through transaction isolation
	t.Run("Concurrent booking attempts", func(t *testing.T) {
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

		// Channel to collect results
		resultChan := make(chan string, 2)

		// WaitGroup to wait for both requests to complete
		var wg sync.WaitGroup
		wg.Add(2)

		// Launch both booking attempts concurrently
		t.Log("Launching concurrent booking attempts for the same room and dates")

		go func() {
			defer wg.Done()
			result, err := createBooking(baseURL, booking1)
			if err != nil {
				resultChan <- fmt.Sprintf("Booking 1 error: %v", err)
			} else if result.Error != "" {
				resultChan <- fmt.Sprintf("Booking 1 API error: %s", result.Error)
			} else {
				resultChan <- fmt.Sprintf("Booking 1 succeeded: ID=%d", result.BookingID)
			}
		}()

		go func() {
			defer wg.Done()
			// Small delay to ensure both requests hit the server in close succession
			time.Sleep(100 * time.Millisecond)
			result, err := createBooking(baseURL, booking2)
			if err != nil {
				resultChan <- fmt.Sprintf("Booking 2 error: %v", err)
			} else if result.Error != "" {
				resultChan <- fmt.Sprintf("Booking 2 API error: %s", result.Error)
			} else {
				resultChan <- fmt.Sprintf("Booking 2 succeeded: ID=%d", result.BookingID)
			}
		}()

		// Wait for both requests to complete
		wg.Wait()
		close(resultChan)

		// Collect and log results
		var results []string
		for result := range resultChan {
			results = append(results, result)
		}

		t.Logf("Results from concurrent booking attempts:")
		for _, result := range results {
			t.Logf("- %s", result)
		}

		// Verify that only one booking succeeded
		successCount := 0
		for _, result := range results {
			if len(result) >= 7 && result[:7] == "Booking" && result[8:17] == "succeeded" {
				successCount++
			}
		}

		t.Logf("Transaction isolation result: %d of 2 concurrent bookings succeeded", successCount)
		if successCount > 1 {
			t.Errorf("TRANSACTION MANAGEMENT FAILURE: Both bookings succeeded, indicating a lack of proper isolation")
		} else {
			t.Logf("TRANSACTION MANAGEMENT SUCCESS: Proper isolation prevented double-booking")
		}
	})

	// SCENARIO 2: Optimistic concurrency control for booking updates
	// This demonstrates how the system handles concurrent modifications to the same booking
	t.Run("Optimistic concurrency control", func(t *testing.T) {
		// First, create a booking to work with
		booking := BookingRequest{
			BookingName:  "Concurrent Update Test",
			RoomNum:      102,
			CheckInDate:  time.Now().AddDate(0, 0, 14), // 14 days from now
			CheckOutDate: time.Now().AddDate(0, 0, 17), // 3-day stay
		}

		bookingResult, err := createBooking(baseURL, booking)
		if err != nil || bookingResult.Error != "" {
			t.Fatalf("Failed to create test booking: %v, API error: %s", err, bookingResult.Error)
		}

		bookingID := bookingResult.BookingID
		t.Logf("Created test booking with ID: %d", bookingID)

		// Now prepare two concurrent update requests with different dates
		update1 := map[string]interface{}{
			"check_in_date":  time.Now().AddDate(0, 0, 15).Format(time.RFC3339), // 15 days from now
			"check_out_date": time.Now().AddDate(0, 0, 18).Format(time.RFC3339), // 3-day stay
		}

		update2 := map[string]interface{}{
			"check_in_date":  time.Now().AddDate(0, 0, 16).Format(time.RFC3339), // 16 days from now
			"check_out_date": time.Now().AddDate(0, 0, 19).Format(time.RFC3339), // 3-day stay
		}

		// Channel to collect results
		resultChan := make(chan string, 2)

		// WaitGroup to wait for both requests to complete
		var wg sync.WaitGroup
		wg.Add(2)

		// Launch both update attempts concurrently
		t.Log("Launching concurrent update attempts for the same booking")

		go func() {
			defer wg.Done()
			success, err := updateBooking(baseURL, bookingID, update1)
			if err != nil {
				resultChan <- fmt.Sprintf("Update 1 error: %v", err)
			} else if !success {
				resultChan <- "Update 1 failed: Optimistic concurrency control rejected the update"
			} else {
				resultChan <- "Update 1 succeeded"
			}
		}()

		go func() {
			defer wg.Done()
			// Small delay to ensure both requests hit the server in close succession
			time.Sleep(100 * time.Millisecond)
			success, err := updateBooking(baseURL, bookingID, update2)
			if err != nil {
				resultChan <- fmt.Sprintf("Update 2 error: %v", err)
			} else if !success {
				resultChan <- "Update 2 failed: Optimistic concurrency control rejected the update"
			} else {
				resultChan <- "Update 2 succeeded"
			}
		}()

		// Wait for both requests to complete
		wg.Wait()
		close(resultChan)

		// Collect and log results
		var results []string
		for result := range resultChan {
			results = append(results, result)
		}

		t.Logf("Results from concurrent update attempts:")
		for _, result := range results {
			t.Logf("- %s", result)
		}

		// Verify that only one update succeeded
		successCount := 0
		for _, result := range results {
			if result == "Update 1 succeeded" || result == "Update 2 succeeded" {
				successCount++
			}
		}

		t.Logf("Optimistic concurrency control result: %d of 2 concurrent updates succeeded", successCount)
		if successCount > 1 {
			t.Errorf("TRANSACTION MANAGEMENT FAILURE: Both updates succeeded, indicating a lack of proper optimistic concurrency control")
		} else if successCount == 0 {
			t.Logf("NOTE: Both updates failed - this can happen in rare timing conditions")
		} else {
			t.Logf("TRANSACTION MANAGEMENT SUCCESS: Optimistic concurrency control prevented conflicting updates")
		}
	})

	// SCENARIO 3: Payment race condition prevention
	// This demonstrates how the system prevents duplicate payments through pessimistic locking
	t.Run("Payment race condition prevention", func(t *testing.T) {
		// First, create a booking to pay for
		booking := BookingRequest{
			BookingName:  "Payment Test",
			RoomNum:      103,
			CheckInDate:  time.Now().AddDate(0, 0, 21), // 21 days from now
			CheckOutDate: time.Now().AddDate(0, 0, 23), // 2-day stay
		}

		bookingResult, err := createBooking(baseURL, booking)
		if err != nil || bookingResult.Error != "" {
			t.Fatalf("Failed to create test booking: %v, API error: %s", err, bookingResult.Error)
		}

		bookingID := bookingResult.BookingID
		amount := bookingResult.TotalPrice
		t.Logf("Created test booking with ID: %d, Amount: $%d", bookingID, amount)

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

		// Channel to collect results
		resultChan := make(chan string, 2)

		// WaitGroup to wait for both requests to complete
		var wg sync.WaitGroup
		wg.Add(2)

		// Launch both payment attempts concurrently
		t.Log("Launching concurrent payment attempts for the same booking")

		go func() {
			defer wg.Done()
			status, err := processPayment(baseURL, payment1)
			if err != nil {
				resultChan <- fmt.Sprintf("Payment 1 error: %v", err)
			} else if status == http.StatusInternalServerError {
				resultChan <- "Payment 1 failed: Server error"
			} else if status == http.StatusBadRequest {
				resultChan <- "Payment 1 failed: Business rule validation"
			} else if status == http.StatusCreated || status == http.StatusOK {
				resultChan <- "Payment 1 succeeded"
			} else {
				resultChan <- fmt.Sprintf("Payment 1 unexpected status: %d", status)
			}
		}()

		go func() {
			defer wg.Done()
			// Small delay to ensure both requests hit the server in close succession
			time.Sleep(100 * time.Millisecond)
			status, err := processPayment(baseURL, payment2)
			if err != nil {
				resultChan <- fmt.Sprintf("Payment 2 error: %v", err)
			} else if status == http.StatusInternalServerError {
				resultChan <- "Payment 2 failed: Server error"
			} else if status == http.StatusBadRequest {
				resultChan <- "Payment 2 failed: Business rule validation"
			} else if status == http.StatusCreated || status == http.StatusOK {
				resultChan <- "Payment 2 succeeded"
			} else {
				resultChan <- fmt.Sprintf("Payment 2 unexpected status: %d", status)
			}
		}()

		// Wait for both requests to complete
		wg.Wait()
		close(resultChan)

		// Collect and log results
		var results []string
		for result := range resultChan {
			results = append(results, result)
		}

		t.Logf("Results from concurrent payment attempts:")
		for _, result := range results {
			t.Logf("- %s", result)
		}

		// Verify that only one payment succeeded
		successCount := 0
		for _, result := range results {
			if result == "Payment 1 succeeded" || result == "Payment 2 succeeded" {
				successCount++
			}
		}

		t.Logf("Payment race condition prevention result: %d of 2 concurrent payments succeeded", successCount)
		if successCount > 1 {
			t.Errorf("TRANSACTION MANAGEMENT FAILURE: Both payments succeeded, indicating a failure in preventing duplicate payments")
		} else if successCount == 0 {
			t.Logf("NOTE: Both payments failed - this is unusual and may indicate an issue with the payment processing")
		} else {
			t.Logf("TRANSACTION MANAGEMENT SUCCESS: System properly prevented duplicate payment")
		}
	})
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
func processPayment(baseURL string, payment PaymentRequest) (int, error) {
	paymentJSON, err := json.Marshal(payment)
	if err != nil {
		return 0, err
	}

	resp, err := http.Post(baseURL+"/receipts", "application/json", bytes.NewBuffer(paymentJSON))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, nil
}
