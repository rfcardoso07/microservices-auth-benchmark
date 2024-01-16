package main

type transferRequestBody struct {
	SenderID   int `json:"senderID" binding:"required"`
	ReceiverID int `json:"receiverID" binding:"required"`
	Amount     int `json:"amount" binding:"required"`
}

type transferAndNotifyRequestBody struct {
	SenderID   int `json:"senderID" binding:"required"`
	ReceiverID int `json:"receiverID" binding:"required"`
	Amount     int `json:"amount" binding:"required"`
}

type getRequestBody struct {
	TransactionID int `json:"transactionID" binding:"required"`
}

type addToAccountRequestPayload struct {
	AccountID int `json:"accountID"`
	Amount    int `json:"amount"`
}

type subtractFromAccountRequestPayload struct {
	AccountID int `json:"accountID"`
	Amount    int `json:"amount"`
}

type notifyRequestPayload struct {
	TransactionID int `json:"transactionID"`
	ReceiverID    int `json:"receiverID"`
	Amount        int `json:"amount"`
}

type authRequestPayload struct {
	UserID    string `json:"userID"`
	Password  string `json:"password"`
	Operation string `json:"operation"`
}

type addToAccountResponseBody struct {
	Message   string `json:"message"`
	AccountID int    `json:"accountID"`
	Amount    int    `json:"amountAdded"`
}

type subtractFromAccountResponseBody struct {
	Message   string `json:"message"`
	AccountID int    `json:"accountID"`
	Amount    int    `json:"amountSubtracted"`
}

type notifyResponseBody struct {
	Message        string `json:"message"`
	NotificationID int    `json:"notificationID"`
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
