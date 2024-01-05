package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"

	"github.com/gin-gonic/gin"
)

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

func (d database) registerNotificationInDatabase(transactionID int, receiverID int, amount int) (int, error) {
	var notificationID int
	// Insert data into the notifications table and retrieve the inserted id
	err := d.DB.QueryRow("INSERT INTO notifications (transaction_id, receiver_id, amount) VALUES ($1, $2, $3) RETURNING notification_id", transactionID, receiverID, amount).Scan(&notificationID)
	return notificationID, err
}

func (d database) getNotificationFromDatabase(notificationID int) (int, int, int, error) {
	// Get transaction data from the transactions table
	var transactionID, receiverID, amount int
	row := d.DB.QueryRow("SELECT transaction_id, receiver_id, amount FROM notifications WHERE id = $1", notificationID)
	err := row.Scan(&transactionID, &receiverID, &amount)
	if err != nil {
		return 0, 0, 0, err
	}
	return transactionID, receiverID, amount, nil
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
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func sendGetAccountRequest(accountID int, accountService string) (getAccountResponseBody, error) {
	payload := getAccountRequestPayload{
		AccountID: accountID,
	}

	// Marshal the struct into a JSON-formatted byte slice
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return getAccountResponseBody{}, err
	}

	url := "http://" + accountService + "/getAccount"

	body, err := performPostRequest(&http.Client{}, url, jsonPayload)
	if err != nil {
		log.Printf("Error performing POST request: %v", err)
		return getAccountResponseBody{}, err
	}

	// Unmarshal the JSON response into a struct
	var response getAccountResponseBody
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Printf("Error unmarshaling JSON: %v", err)
		return getAccountResponseBody{}, err
	}

	return response, nil
}

func sendGetCustomerRequest(customerID int, customerService string) (getCustomerResponseBody, error) {
	payload := getCustomerRequestPayload{
		CustomerID: customerID,
	}

	// Marshal the struct into a JSON-formatted byte slice
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return getCustomerResponseBody{}, err
	}

	url := "http://" + customerService + "/getCustomer"

	body, err := performPostRequest(&http.Client{}, url, jsonPayload)
	if err != nil {
		log.Printf("Error performing POST request: %v", err)
		return getCustomerResponseBody{}, err
	}

	// Unmarshal the JSON response into a struct
	var response getCustomerResponseBody
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Printf("Error unmarshaling JSON: %v", err)
		return getCustomerResponseBody{}, err
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
		Host:     os.Getenv("NOTIFICATION_SERVICE_DATABASE_HOST"),
		Port:     os.Getenv("NOTIFICATION_SERVICE_DATABASE_PORT"),
		User:     os.Getenv("NOTIFICATION_SERVICE_DATABASE_USER"),
		Password: os.Getenv("NOTIFICATION_SERVICE_DATABASE_PASSWORD"),
		Name:     os.Getenv("NOTIFICATION_SERVICE_DATABASE_NAME"),
		DB:       &sql.DB{},
	}

	err := d.init()
	if err != nil {
		return
	}

	customerService := os.Getenv("CUSTOMER_SERVICE_HOST_AND_PORT")
	accountService := os.Getenv("ACCOUNT_SERVICE_HOST_AND_PORT")
	authService := os.Getenv("AUTH_SERVICE_HOST_AND_PORT")
	authPattern := os.Getenv("APPLICATION_AUTH_PATTERN")

	// Create a new Gin router
	r := gin.Default()

	// Route for sending notifications
	r.POST("/notify", func(c *gin.Context) {
		var requestBody notifyRequestBody

		// Bind the JSON body to the RequestBody struct
		if err := c.BindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		getAccountResponse, err := sendGetAccountRequest(requestBody.ReceiverID, accountService)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		getCustomerResponse, err := sendGetCustomerRequest(getAccountResponse.CustomerID, customerService)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		notificationID, err := d.registerNotificationInDatabase(requestBody.TransactionID, requestBody.ReceiverID, requestBody.Amount)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		log.Println("This should be an e-mail trigger, but for now it is only a log message.")

		c.JSON(http.StatusOK, gin.H{
			"message":        "success",
			"notificationID": notificationID,
			"recipientEmail": getCustomerResponse.CustomerEmail,
		})
	})

	// Route for retrieving notifications data
	r.POST("/getNotification", func(c *gin.Context) {
		var requestBody getRequestBody

		// Bind the JSON body to the RequestBody struct
		if err := c.BindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		transactionID, receiverID, amount, err := d.getNotificationFromDatabase(requestBody.NotificationID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":        "success",
			"notificationID": requestBody.NotificationID,
			"transactionID":  transactionID,
			"receiverID":     receiverID,
			"amount":         amount,
		})
	})

	// Run the server on port 8086
	r.Run(":8086")
	defer d.DB.Close()
}
