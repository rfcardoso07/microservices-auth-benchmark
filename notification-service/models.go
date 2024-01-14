package main

type notifyRequestBody struct {
	TransactionID int `json:"transactionID" binding:"required"`
	ReceiverID    int `json:"receiverID" binding:"required"`
	Amount        int `json:"amount" binding:"required"`
}

type getRequestBody struct {
	NotificationID int `json:"notificationID" binding:"required"`
}

type getAccountRequestPayload struct {
	AccountID int `json:"accountID"`
}

type getCustomerRequestPayload struct {
	CustomerID int `json:"customerID"`
}

type authRequestPayload struct {
	UserID    string `json:"userID"`
	Password  string `json:"password"`
	Operation string `json:"operation"`
}

type getAccountResponseBody struct {
	Message    string `json:"message"`
	AccountID  int    `json:"accountID"`
	CustomerID int    `json:"customerID"`
	Balance    int    `json:"balance"`
}

type getCustomerResponseBody struct {
	Message       string `json:"message"`
	CustomerID    int    `json:"customerID"`
	CustomerName  string `json:"customerName"`
	CustomerEmail string `json:"customerEmail"`
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
