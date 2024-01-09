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
	"time"

	_ "github.com/lib/pq"

	"github.com/gin-gonic/gin"
)

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

func (d database) registerBalanceInDatabase(customerID int, totalBalance int) error {
	currentTime := time.Now()
	// Insert data into the balances table
	_, err := d.DB.Exec("INSERT INTO balances (customer_id, total_balance, registered_at) VALUES ($1, $2, $3)", totalBalance, currentTime)
	return err
}

func (d database) getLatestRecordsFromDatabase(customerID int, numberOfRecords int) ([]int, []time.Time, error) {
	// Retrieve the latest entries for the customer ID
	rows, err := d.DB.Query("SELECT total_balance, registered_at FROM balances WHERE customer_id = $1 ORDER BY registered_at DESC LIMIT $2", customerID, numberOfRecords)
	if err != nil {
		return []int{}, []time.Time{}, err
	}
	defer rows.Close()

	var totalBalances []int
	var registeredAt []time.Time

	// Iterate through the rows and add entries to the slices
	for rows.Next() {
		var balance int
		var timestamp time.Time
		if err := rows.Scan(&balance, &timestamp); err != nil {
			return []int{}, []time.Time{}, err
		}
		totalBalances = append(totalBalances, balance)
		registeredAt = append(registeredAt, timestamp)
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		return []int{}, []time.Time{}, err
	}

	return totalBalances, registeredAt, nil
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

func sendGetAccountsByCustomerRequest(customerID int, accountService string, userID string, userPassword string) (getAccountsByCustomerResponseBody, error) {
	payload := getAccountsByCustomerRequestPayload{
		CustomerID: customerID,
	}

	// Marshal the struct into a JSON-formatted byte slice
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return getAccountsByCustomerResponseBody{}, err
	}

	url := "http://" + accountService + "/getAccountsByCustomer"
	if userID != "" && userPassword != "" {
		url = url + "/" + userID + "/" + userPassword
	}

	body, err := performPostRequest(&http.Client{}, url, jsonPayload)
	if err != nil {
		log.Printf("Error performing POST request: %v", err)
		return getAccountsByCustomerResponseBody{}, err
	}

	// Unmarshal the JSON response into a struct
	var response getAccountsByCustomerResponseBody
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Printf("Error unmarshaling JSON: %v", err)
		return getAccountsByCustomerResponseBody{}, err
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
		Host:     os.Getenv("BALANCE_SERVICE_DATABASE_HOST"),
		Port:     os.Getenv("BALANCE_SERVICE_DATABASE_PORT"),
		User:     os.Getenv("BALANCE_SERVICE_DATABASE_USER"),
		Password: os.Getenv("BALANCE_SERVICE_DATABASE_PASSWORD"),
		Name:     os.Getenv("BALANCE_SERVICE_DATABASE_NAME"),
		DB:       &sql.DB{},
	}

	err := d.init()
	if err != nil {
		return
	}

	accountService := os.Getenv("ACCOUNT_SERVICE_HOST_AND_PORT")
	authService := os.Getenv("AUTH_SERVICE_HOST_AND_PORT")
	authPattern := os.Getenv("APPLICATION_AUTH_PATTERN")

	// Create a new Gin router
	r := gin.Default()

	switch authPattern {
	case "NO_AUTH":
		// Route for getting customers total balance
		r.POST("/getBalanceByCustomer", func(c *gin.Context) {
			var requestBody getRequestBody

			// Bind the JSON body to the RequestBody struct
			if err := c.BindJSON(&requestBody); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			response, err := sendGetAccountsByCustomerRequest(requestBody.CustomerID, accountService, "", "")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			totalBalance := 0
			for _, balance := range response.Balances {
				totalBalance += balance
			}

			err = d.registerBalanceInDatabase(requestBody.CustomerID, totalBalance)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"message":      "success",
				"customerID":   requestBody.CustomerID,
				"accountIDs":   response.AccountIDs,
				"balances":     response.Balances,
				"totalBalance": totalBalance,
			})
		})

		// Route for deleting customers
		r.POST("/getBalanceHistory", func(c *gin.Context) {
			var requestBody getHistoryRequestBody

			// Bind the JSON body to the RequestBody struct
			if err := c.BindJSON(&requestBody); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			totalBalances, timestamps, err := d.getLatestRecordsFromDatabase(requestBody.CustomerID, requestBody.NumberOfRecords)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"message":          "success",
				"customerID":       requestBody.CustomerID,
				"totalBalances":    totalBalances,
				"recordTimestamps": timestamps,
			})
		})

	case "CENTRALIZED":
		// Route for getting customers total balance
		r.POST("/getBalanceByCustomer/:id/:password", func(c *gin.Context) {
			userID := c.Param("id")
			userPassword := c.Param("password")

			authResponse, err := sendAuthRequest(userID, userPassword, "READ", authService)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			if authResponse.AccessGranted {
				var requestBody getRequestBody

				// Bind the JSON body to the RequestBody struct
				if err := c.BindJSON(&requestBody); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				response, err := sendGetAccountsByCustomerRequest(requestBody.CustomerID, accountService, userID, userPassword)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				totalBalance := 0
				for _, balance := range response.Balances {
					totalBalance += balance
				}

				err = d.registerBalanceInDatabase(requestBody.CustomerID, totalBalance)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"message":      "success",
					"customerID":   requestBody.CustomerID,
					"accountIDs":   response.AccountIDs,
					"balances":     response.Balances,
					"totalBalance": totalBalance,
				})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"message":       "accessDenied",
					"authenticated": authResponse.Authenticated,
					"authorized":    authResponse.Authorized,
				})
			}
		})

		// Route for deleting customers
		r.POST("/getBalanceHistory/:id/:password", func(c *gin.Context) {
			userID := c.Param("id")
			userPassword := c.Param("password")

			authResponse, err := sendAuthRequest(userID, userPassword, "READ", authService)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			if authResponse.AccessGranted {
				var requestBody getHistoryRequestBody

				// Bind the JSON body to the RequestBody struct
				if err := c.BindJSON(&requestBody); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				totalBalances, timestamps, err := d.getLatestRecordsFromDatabase(requestBody.CustomerID, requestBody.NumberOfRecords)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"message":          "success",
					"customerID":       requestBody.CustomerID,
					"totalBalances":    totalBalances,
					"recordTimestamps": timestamps,
				})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"message":       "accessDenied",
					"authenticated": authResponse.Authenticated,
					"authorized":    authResponse.Authorized,
				})
			}
		})

	case "DECENTRALIZED":
		// Route for getting customers total balance
		r.POST("/getBalanceByCustomer/:id/:password", func(c *gin.Context) {
			userID := c.Param("id")
			userPassword := c.Param("password")

			authenticated, authorized, accessGranted, err := d.authenticateAndAuthorize(userID, userPassword, "READ")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			if accessGranted {
				var requestBody getRequestBody

				// Bind the JSON body to the RequestBody struct
				if err := c.BindJSON(&requestBody); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				response, err := sendGetAccountsByCustomerRequest(requestBody.CustomerID, accountService, userID, userPassword)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				totalBalance := 0
				for _, balance := range response.Balances {
					totalBalance += balance
				}

				err = d.registerBalanceInDatabase(requestBody.CustomerID, totalBalance)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"message":      "success",
					"customerID":   requestBody.CustomerID,
					"accountIDs":   response.AccountIDs,
					"balances":     response.Balances,
					"totalBalance": totalBalance,
				})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"message":       "accessDenied",
					"authenticated": authenticated,
					"authorized":    authorized,
				})
			}
		})

		// Route for deleting customers
		r.POST("/getBalanceHistory/:id/:password", func(c *gin.Context) {
			userID := c.Param("id")
			userPassword := c.Param("password")

			authenticated, authorized, accessGranted, err := d.authenticateAndAuthorize(userID, userPassword, "READ")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			if accessGranted {
				var requestBody getHistoryRequestBody

				// Bind the JSON body to the RequestBody struct
				if err := c.BindJSON(&requestBody); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				totalBalances, timestamps, err := d.getLatestRecordsFromDatabase(requestBody.CustomerID, requestBody.NumberOfRecords)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"message":          "success",
					"customerID":       requestBody.CustomerID,
					"totalBalances":    totalBalances,
					"recordTimestamps": timestamps,
				})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"message":       "accessDenied",
					"authenticated": authenticated,
					"authorized":    authorized,
				})
			}
		})

	default:
		log.Printf("Unexpected APPLICATION_AUTH_PATTERN: %v", authPattern)
		return
	}

	// Run the server on port 8088
	r.Run(":8088")
	defer d.DB.Close()
}
