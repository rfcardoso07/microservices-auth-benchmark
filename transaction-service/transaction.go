package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"

	"github.com/gin-gonic/gin"
)

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
	Amount        int `json:"amount"`
	ReceiverID    int `json:"receiverID"`
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

type database struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	DB       *sql.DB
}

func (d *database) init() error {
	// Create connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		d.Host, d.Port, d.User, d.Password, d.Name)

	// Open a database connection and set up a connection pool
	var err error
	d.DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("Error opening database connection: %v", err)
		return err
	}

	// Set the maximum number of open (in-use + idle) connections
	d.DB.SetMaxOpenConns(10)

	// Set the maximum number of idle connections in the pool
	d.DB.SetMaxIdleConns(5)

	// Check if the database connection is alive
	err = d.DB.Ping()
	if err != nil {
		log.Printf("Error pinging database: %v", err)
		return err
	}

	log.Println("Connected to the database")
	return nil
}

func (d database) createTransactionInDatabase(senderID int, receiverID int, amount int) (int, error) {
	var transactionID int
	// Insert data into the transactions table and retrieve the inserted id
	err := d.DB.QueryRow("INSERT INTO transactions (sender_id, receiver_id, amount) VALUES ($1, $2, $3) RETURNING transaction_id", senderID, receiverID, amount).Scan(&transactionID)
	return transactionID, err
}

func (d database) getTransactionFromDatabase(transactionID int) (int, int, int, error) {
	// Get transaction data from the transactions table
	var senderID, receiverID, amount int
	row := d.DB.QueryRow("SELECT sender_id, receiver_id, amount FROM transactions WHERE transaction_id = $1", transactionID)
	err := row.Scan(&senderID, &receiverID, &amount)
	if err != nil {
		return 0, 0, 0, err
	}
	return senderID, receiverID, amount, nil
}

func (d database) searchForUserInDatabase(userID string, password string) (bool, userPermissions, error) {
	// Search for userID and password in the users table and retrieve permissions
	var permissions userPermissions
	row := d.DB.QueryRow("SELECT can_read, can_write, can_delete FROM users WHERE user_id = $1 AND user_password = $2", userID, password)
	err := row.Scan(&permissions.CanRead, &permissions.CanWrite, &permissions.CanDelete)

	if err != nil {
		if err == sql.ErrNoRows {
			// Not actually an error, just means there was no match (user + password)
			return false, userPermissions{}, nil
		} else {
			return false, userPermissions{}, err
		}
	}

	return true, permissions, nil
}

func hasPermission(operation string, permissions userPermissions) bool {
	switch operation {
	case "READ":
		if permissions.CanRead {
			return true
		}
		return false
	case "WRITE":
		if permissions.CanWrite {
			return true
		}
		return false
	case "DELETE":
		if permissions.CanDelete {
			return true
		}
		return false
	default:
		return false
	}
}

func (d database) authenticateAndAuthorize(userID string, userPassword string, operation string) (bool, bool, bool, error) {
	authorized := false
	accessGranted := false

	authenticated, permissions, err := d.searchForUserInDatabase(userID, userPassword)
	if err != nil {
		return false, false, false, err
	}

	if authenticated {
		authorized = hasPermission(operation, permissions)
		if authorized {
			accessGranted = true
		}
	}

	return authenticated, authorized, accessGranted, nil
}

func performPostRequest(client *http.Client, url string, payload []byte) ([]byte, error) {
	// Create a POST request with the JSON payload
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	// Set headers
	request.Header.Set("Content-Type", "application/json")

	// Send the request using the provided http.Client
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func sendAddToAccountRequest(accountID int, amount int, accountService string) (addToAccountResponseBody, error) {
	payload := addToAccountRequestPayload{
		AccountID: accountID,
		Amount:    amount,
	}

	// Marshal the struct into a JSON-formatted byte slice
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return addToAccountResponseBody{}, err
	}

	url := "http://" + accountService + "/addToBalance"

	body, err := performPostRequest(&http.Client{}, url, jsonPayload)
	if err != nil {
		log.Printf("Error performing POST request: %v", err)
		return addToAccountResponseBody{}, err
	}

	// Unmarshal the JSON response into a struct
	var response addToAccountResponseBody
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Printf("Error unmarshaling JSON: %v", err)
		return addToAccountResponseBody{}, err
	}

	return response, nil
}

func sendSubtractFromAccountRequest(accountID int, amount int, accountService string) (subtractFromAccountResponseBody, error) {
	payload := subtractFromAccountRequestPayload{
		AccountID: accountID,
		Amount:    amount,
	}

	// Marshal the struct into a JSON-formatted byte slice
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return subtractFromAccountResponseBody{}, err
	}

	url := "http://" + accountService + "/subtractFromBalance"

	body, err := performPostRequest(&http.Client{}, url, jsonPayload)
	if err != nil {
		log.Printf("Error performing POST request: %v", err)
		return subtractFromAccountResponseBody{}, err
	}

	// Unmarshal the JSON response into a struct
	var response subtractFromAccountResponseBody
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Printf("Error unmarshaling JSON: %v", err)
		return subtractFromAccountResponseBody{}, err
	}

	return response, nil
}

func sendNotifyRequest(transactionID int, amount int, receiverID int, notificationService string) (notifyResponseBody, error) {
	payload := notifyRequestPayload{
		TransactionID: transactionID,
		Amount:        amount,
		ReceiverID:    receiverID,
	}

	// Marshal the struct into a JSON-formatted byte slice
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return notifyResponseBody{}, err
	}

	url := "http://" + notificationService + "/notify"

	body, err := performPostRequest(&http.Client{}, url, jsonPayload)
	if err != nil {
		log.Printf("Error performing POST request: %v", err)
		return notifyResponseBody{}, err
	}

	// Unmarshal the JSON response into a struct
	var response notifyResponseBody
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Printf("Error unmarshaling JSON: %v", err)
		return notifyResponseBody{}, err
	}

	return response, nil
}

func sendAuthRequest(userID string, userPassword string, operation string, authService string) (authResponseBody, error) {
	payload := authRequestPayload{
		UserID:    userID,
		Password:  userPassword,
		Operation: operation,
	}

	// Marshal the struct into a JSON-formatted byte slice
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return authResponseBody{}, err
	}

	url := "http://" + authService + "/authenticateAndAuthorize"

	body, err := performPostRequest(&http.Client{}, url, jsonPayload)
	if err != nil {
		log.Printf("Error performing POST request: %v", err)
		return authResponseBody{}, err
	}

	// Unmarshal the JSON response into a struct
	var response authResponseBody
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Printf("Error unmarshaling JSON: %v", err)
		return authResponseBody{}, err
	}

	return response, nil
}

func main() {
	gin.SetMode(gin.DebugMode)

	d := database{
		Host:     os.Getenv("TRANSACTION_SERVICE_DATABASE_HOST"),
		Port:     os.Getenv("TRANSACTION_SERVICE_DATABASE_PORT"),
		User:     os.Getenv("TRANSACTION_SERVICE_DATABASE_USER"),
		Password: os.Getenv("TRANSACTION_SERVICE_DATABASE_PASSWORD"),
		Name:     os.Getenv("TRANSACTION_SERVICE_DATABASE_NAME"),
		DB:       &sql.DB{},
	}

	err := d.init()
	if err != nil {
		return
	}

	accountService := os.Getenv("ACCOUNT_SERVICE_HOST_AND_PORT")
	notificationService := os.Getenv("NOTIFICATION_SERVICE_HOST_AND_PORT")
	authService := os.Getenv("AUTH_SERVICE_HOST_AND_PORT")
	authPattern := os.Getenv("APPLICATION_AUTH_PATTERN")

	// Create a new Gin router
	r := gin.Default()

	// Route for performing transactions
	r.POST("/transferAmount", func(c *gin.Context) {
		var requestBody transferRequestBody

		// Bind the JSON body to the RequestBody struct
		if err := c.BindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		subtractResponse, err := sendSubtractFromAccountRequest(requestBody.SenderID, requestBody.Amount, accountService)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		addResponse, err := sendAddToAccountRequest(requestBody.ReceiverID, requestBody.Amount, accountService)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		transactionID, err := d.createTransactionInDatabase(requestBody.SenderID, requestBody.ReceiverID, requestBody.Amount)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":       "success",
			"transactionID": transactionID,
			"senderID":      subtractResponse.AccountID,
			"receiverID":    addResponse.AccountID,
		})
	})

	// Route for performing transactions and notifying receivers
	r.POST("/transferAmountAndNotify", func(c *gin.Context) {
		var requestBody transferAndNotifyRequestBody

		// Bind the JSON body to the RequestBody struct
		if err := c.BindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		subtractResponse, err := sendSubtractFromAccountRequest(requestBody.SenderID, requestBody.Amount, accountService)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		addResponse, err := sendAddToAccountRequest(requestBody.ReceiverID, requestBody.Amount, accountService)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		transactionID, err := d.createTransactionInDatabase(requestBody.SenderID, requestBody.ReceiverID, requestBody.Amount)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		notifyResponse, err := sendNotifyRequest(transactionID, requestBody.ReceiverID, requestBody.Amount, notificationService)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":        "success",
			"transactionID":  transactionID,
			"senderID":       subtractResponse.AccountID,
			"receiverID":     addResponse.AccountID,
			"notificationID": notifyResponse.NotificationID,
		})
	})

	// Route for retrieving transactions data
	r.POST("/getTransaction", func(c *gin.Context) {
		var requestBody getRequestBody

		// Bind the JSON body to the RequestBody struct
		if err := c.BindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		senderID, receiverID, amount, err := d.getTransactionFromDatabase(requestBody.TransactionID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":       "success",
			"transactionID": requestBody.TransactionID,
			"senderID":      senderID,
			"receiverID":    receiverID,
			"amount":        amount,
		})
	})

	// Run the server on port 8084
	r.Run(":8084")
	defer d.DB.Close()
}
