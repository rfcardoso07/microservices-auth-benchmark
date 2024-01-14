package main

type getRequestBody struct {
	CustomerID int `json:"customerID" binding:"required"`
}

type getHistoryRequestBody struct {
	CustomerID      int `json:"customerID" binding:"required"`
	NumberOfRecords int `json:"numberOfRecords" binding:"required"`
}

type getAccountsByCustomerRequestPayload struct {
	CustomerID int `json:"customerID"`
}

type authRequestPayload struct {
	UserID    string `json:"userID"`
	Password  string `json:"password"`
	Operation string `json:"operation"`
}

type getAccountsByCustomerResponseBody struct {
	Message    string `json:"message"`
	CustomerID int    `json:"customerID"`
	AccountIDs []int  `json:"accountIDs"`
	Balances   []int  `json:"balances"`
}

type authResponseBody struct {
	Message       string `json:"message"`
	Authenticated bool   `json:"authenticated"`
	Authorized    bool   `json:"authorized"`
	AccessGranted bool   `json:"accessGranted"`
}

type userPermissions struct {
	CanRead   bool
	CanWrite  bool
	CanDelete bool
}
