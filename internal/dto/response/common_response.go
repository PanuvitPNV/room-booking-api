package response

// PaginationResponse common pagination structure
type PaginationResponse struct {
	Total    int `json:"total"`
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	LastPage int `json:"last_page"`
}

// ErrorResponse common error structure
type ErrorResponse struct {
	Error       string `json:"error"`
	Code        int    `json:"code"`
	Description string `json:"description,omitempty"`
}

// SuccessResponse common success structure
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// TransactionResponse for transaction results
type TransactionResponse struct {
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	TransactionID string `json:"transaction_id,omitempty"`
}
