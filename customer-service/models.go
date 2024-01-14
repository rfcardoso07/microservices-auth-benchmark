package main

type createRequestBody struct {
	CustomerName  string `json:"customerName" binding:"required"`
	CustomerEmail string `json:"customerEmail" binding:"required"`
}

type deleteRequestBody struct {
	CustomerID int `json:"customerID" binding:"required"`
}

type getRequestBody struct {
	CustomerID int `json:"customerID" binding:"required"`
}

type createAccountRequestPayload struct {
	CustomerID int `json:"customerID"`
}

type deleteAccountsByCustomerRequestPayload struct {
	CustomerID int `json:"customerID"`
}

type authRequestPayload struct {
	UserID    string `json:"userID"`
	Password  string `json:"password"`
	Operation string `json:"operation"`
}

type createAccountResponseBody struct {
	Message   string `json:"message"`
	AccountID int    `json:"accountID"`
}

type deleteAccountsByCustomerResponseBody struct {
	Message    string `json:"message"`
	AccountIDs []int  `json:"accountIDs"`
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
