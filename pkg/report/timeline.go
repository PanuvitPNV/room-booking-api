package report

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// TransactionEvent represents a single transaction event for the timeline
type TransactionEvent struct {
	ID           string    `json:"id"`
	ClientID     int       `json:"clientId"`
	ActionType   string    `json:"actionType"`
	BookingID    int       `json:"bookingId,omitempty"`
	RoomNum      int       `json:"roomNum,omitempty"`
	ResourceID   int       `json:"resourceId,omitempty"` // Generic ID for any resource
	StartTime    time.Time `json:"startTime"`
	EndTime      time.Time `json:"endTime"`
	Duration     int64     `json:"duration"` // in milliseconds
	Status       string    `json:"status"`   // success, failed, conflict
	ErrorMessage string    `json:"errorMessage,omitempty"`
	Details      string    `json:"details,omitempty"`
	ResponseTime int64     `json:"responseTime,omitempty"` // in milliseconds
}

// ClientAction from your test code (reference)
type ClientAction struct {
	ActionType string
	Data       interface{}
	Result     interface{}
	Error      error
	StartTime  time.Time
	EndTime    time.Time
}

// TimelineReportData contains all data needed for the timeline report
type TimelineReportData struct {
	TestName          string                 `json:"testName"`
	TestStartTime     time.Time              `json:"testStartTime"`
	TestEndTime       time.Time              `json:"testEndTime"`
	TestDuration      time.Duration          `json:"testDuration"`
	TotalRequests     int                    `json:"totalRequests"`
	SuccessfulRequest int                    `json:"successfulRequests"`
	FailedRequests    int                    `json:"failedRequests"`
	ConflictEvents    int                    `json:"conflictEvents"`
	Events            []*TransactionEvent    `json:"events"`
	ClientStats       map[int]ClientStats    `json:"clientStats"`
	RoomStats         map[int]RoomStats      `json:"roomStats"`
	ActionStats       map[string]ActionStats `json:"actionStats"`
	TimelineJSON      template.JS            `json:"timelineJSON"` // Changed to template.JS
}

// ClientStats tracks statistics for each client
type ClientStats struct {
	TotalActions      int                `json:"totalActions"`
	SuccessfulActions int                `json:"successfulActions"`
	FailedActions     int                `json:"failedActions"`
	ActionCounts      map[string]int     `json:"actionCounts"`
	AverageResponse   int64              `json:"averageResponse"` // in milliseconds
	ResponseByAction  map[string][]int64 `json:"responseByAction"`
}

// RoomStats tracks statistics for each room
type RoomStats struct {
	TotalTransactions int                 `json:"totalTransactions"`
	BookingEvents     int                 `json:"bookingEvents"`
	ConflictEvents    int                 `json:"conflictEvents"`
	SuccessRate       float64             `json:"successRate"`
	ClientsAttempted  map[int]bool        `json:"clientsAttempted"`
	Timeline          []*TransactionEvent `json:"timeline"`
}

// ActionStats tracks statistics for each action type
type ActionStats struct {
	TotalAttempts      int     `json:"totalAttempts"`
	SuccessfulAttempts int     `json:"successfulAttempts"`
	FailedAttempts     int     `json:"failedAttempts"`
	AverageResponse    int64   `json:"averageResponse"` // in milliseconds
	SuccessRate        float64 `json:"successRate"`
}

// TimelineGenerator generates timeline reports from test data
type TimelineGenerator struct {
	reportData   *TimelineReportData
	templatePath string
}

// NewTimelineGenerator creates a new timeline generator
func NewTimelineGenerator(testName string) *TimelineGenerator {
	return &TimelineGenerator{
		reportData: &TimelineReportData{
			TestName:      testName,
			TestStartTime: time.Now(),
			Events:        make([]*TransactionEvent, 0),
			ClientStats:   make(map[int]ClientStats),
			RoomStats:     make(map[int]RoomStats),
			ActionStats:   make(map[string]ActionStats),
		},
		templatePath: "ui/report/timeline.html", // Default template path
	}
}

// SetTemplatePath sets a custom template path
func (g *TimelineGenerator) SetTemplatePath(path string) {
	g.templatePath = path
}

// AddEvent adds a transaction event to the timeline
func (g *TimelineGenerator) AddEvent(event *TransactionEvent) {
	g.reportData.Events = append(g.reportData.Events, event)

	// Update client stats
	clientStats, exists := g.reportData.ClientStats[event.ClientID]
	if !exists {
		clientStats = ClientStats{
			ActionCounts:     make(map[string]int),
			ResponseByAction: make(map[string][]int64),
		}
	}

	clientStats.TotalActions++
	clientStats.ActionCounts[event.ActionType]++

	// Calculate response time
	responseTime := event.EndTime.Sub(event.StartTime).Milliseconds()
	clientStats.ResponseByAction[event.ActionType] = append(
		clientStats.ResponseByAction[event.ActionType],
		responseTime,
	)

	if event.Status == "success" {
		clientStats.SuccessfulActions++
	} else {
		clientStats.FailedActions++
	}

	g.reportData.ClientStats[event.ClientID] = clientStats

	// Update room stats if applicable
	if event.RoomNum > 0 {
		roomStats, exists := g.reportData.RoomStats[event.RoomNum]
		if !exists {
			roomStats = RoomStats{
				ClientsAttempted: make(map[int]bool),
				Timeline:         make([]*TransactionEvent, 0),
			}
		}

		roomStats.TotalTransactions++
		roomStats.ClientsAttempted[event.ClientID] = true
		roomStats.Timeline = append(roomStats.Timeline, event)

		if strings.Contains(event.ActionType, "booking") {
			roomStats.BookingEvents++
		}

		if event.Status == "conflict" {
			roomStats.ConflictEvents++
		}

		if roomStats.TotalTransactions > 0 {
			roomStats.SuccessRate = float64(roomStats.TotalTransactions-roomStats.ConflictEvents) /
				float64(roomStats.TotalTransactions)
		}

		g.reportData.RoomStats[event.RoomNum] = roomStats
	}

	// Update action stats
	actionStats, exists := g.reportData.ActionStats[event.ActionType]
	if !exists {
		actionStats = ActionStats{}
	}

	actionStats.TotalAttempts++

	if event.Status == "success" {
		actionStats.SuccessfulAttempts++
	} else {
		actionStats.FailedAttempts++
	}

	if actionStats.TotalAttempts > 0 {
		actionStats.SuccessRate = float64(actionStats.SuccessfulAttempts) /
			float64(actionStats.TotalAttempts)
	}

	// Update average response time
	totalResponseTime := actionStats.AverageResponse * int64(actionStats.TotalAttempts-1)
	totalResponseTime += event.EndTime.Sub(event.StartTime).Milliseconds()
	actionStats.AverageResponse = totalResponseTime / int64(actionStats.TotalAttempts)

	g.reportData.ActionStats[event.ActionType] = actionStats

	// Update global stats
	g.reportData.TotalRequests++
	if event.Status == "success" {
		g.reportData.SuccessfulRequest++
	} else if event.Status == "conflict" {
		g.reportData.ConflictEvents++
		g.reportData.FailedRequests++
	} else if event.Status == "failed" {
		g.reportData.FailedRequests++
	}
}

// ConvertClientActionToEvent converts a client action to a timeline event
func (g *TimelineGenerator) ConvertClientActionToEvent(clientID int, action *ClientAction, booking interface{}, room int) *TransactionEvent {
	event := &TransactionEvent{
		ID:         fmt.Sprintf("evt-%d-%s-%d", clientID, action.ActionType, time.Now().UnixNano()),
		ClientID:   clientID,
		ActionType: action.ActionType,
		StartTime:  action.StartTime,
		EndTime:    action.EndTime,
		Duration:   action.EndTime.Sub(action.StartTime).Milliseconds(),
		RoomNum:    room,
		Status:     "success",
	}

	// If there's an error, set the status and message
	if action.Error != nil {
		event.Status = "failed"
		event.ErrorMessage = action.Error.Error()

		if strings.Contains(event.ErrorMessage, "conflict") ||
			strings.Contains(event.ErrorMessage, "already booked") ||
			strings.Contains(event.ErrorMessage, "unavailable") {
			event.Status = "conflict"
		}
	}

	// Add booking-specific details if available
	if booking != nil {
		bookingJSON, _ := json.Marshal(booking)
		event.Details = string(bookingJSON)

		// Extract booking ID and room number if available
		// This will depend on your booking structure
		if b, ok := booking.(map[string]interface{}); ok {
			if id, exists := b["booking_id"]; exists {
				if bookingID, ok := id.(float64); ok {
					event.BookingID = int(bookingID)
				}
			}
			if roomNum, exists := b["room_num"]; exists {
				if rn, ok := roomNum.(float64); ok {
					event.RoomNum = int(rn)
				}
			}
		}
	}

	return event
}

// FinishTest completes the test and calculates final statistics
func (g *TimelineGenerator) FinishTest() {
	g.reportData.TestEndTime = time.Now()
	g.reportData.TestDuration = g.reportData.TestEndTime.Sub(g.reportData.TestStartTime)

	// Sort events by start time
	sort.Slice(g.reportData.Events, func(i, j int) bool {
		return g.reportData.Events[i].StartTime.Before(g.reportData.Events[j].StartTime)
	})

	// Calculate client averages
	for clientID, stats := range g.reportData.ClientStats {
		var totalResponse int64
		var count int64

		for _, responses := range stats.ResponseByAction {
			for _, respTime := range responses {
				totalResponse += respTime
				count++
			}
		}

		if count > 0 {
			stats.AverageResponse = totalResponse / count
		}

		g.reportData.ClientStats[clientID] = stats
	}

	// Convert events to JSON for visualization
	// THIS IS THE KEY FIX: properly format the JSON as a string that can be parsed by JavaScript
	// eventsJSON, err := json.Marshal(g.reportData.Events)
	// if err != nil {
	// 	log.Printf("Error marshaling events: %v", err)
	// 	g.reportData.TimelineJSON = template.JS("[]") // Empty array as fallback
	// } else {
	// 	g.reportData.TimelineJSON = template.JS(string(eventsJSON)) // Properly formatted
	// }
}

// Update the GenerateReport method to properly include template functions
func (g *TimelineGenerator) GenerateReport(outputPath string) error {
	// Ensure output directory exists
	err := os.MkdirAll(filepath.Dir(outputPath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create template functions map
	funcMap := template.FuncMap{
		"percentage": func(part, total int) float64 {
			if total == 0 {
				return 0
			}
			return float64(part) / float64(total) * 100
		},
		"multiply": func(a, b float64) float64 {
			return a * b
		},
		"topActions": func(actions map[string]int, n int) map[string]int {
			type kv struct {
				Key   string
				Value int
			}

			if len(actions) <= n {
				return actions
			}

			var sorted []kv
			for k, v := range actions {
				sorted = append(sorted, kv{k, v})
			}

			sort.Slice(sorted, func(i, j int) bool {
				return sorted[i].Value > sorted[j].Value
			})

			result := make(map[string]int)
			for i := 0; i < n && i < len(sorted); i++ {
				result[sorted[i].Key] = sorted[i].Value
			}

			return result
		},
		// Add this function to safely convert data to JSON
		"json": func(v interface{}) template.JS {
			b, err := json.Marshal(v)
			if err != nil {
				return template.JS("[]")
			}
			return template.JS(string(b))
		},
	}

	// Read template file
	tmpl, err := template.New(filepath.Base(g.templatePath)).
		Funcs(funcMap).
		ParseFiles(g.templatePath)

	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Create output file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Execute template with data
	err = tmpl.Execute(file, g.reportData)
	if err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	fmt.Printf("Timeline report generated at: %s\n", outputPath)

	return nil
}
