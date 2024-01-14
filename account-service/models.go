package main

type createRequestBody struct {
	CustomerID int `json:"customerID" binding:"required"`
}

type deleteRequestBody struct {
	AccountID int `json:"accountID" binding:"required"`
}

type deleteByCustomerRequestBody struct {
	CustomerID int `json:"customerID" binding:"required"`
}

type getRequestBody struct {
	AccountID int `json:"accountID" binding:"required"`
}

type getByCustomerRequestBody struct {
	CustomerID int `json:"customerID" binding:"required"`
}

type addToBalanceRequestBody struct {
	AccountID int `json:"accountID" binding:"required"`
	Amount    int `json:"amount" binding:"required"`
}

type subtractFromBalanceRequestBody struct {
	AccountID int `json:"accountID" binding:"required"`
	Amount    int `json:"amount" binding:"required"`
}

type authRequestPayload struct {
	UserID    string `json:"userID"`
	Password  string `json:"password"`
	Operation string `json:"operation"`
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
