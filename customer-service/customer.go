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

type createRequestBody struct {
	Name  string `json:"customerName" binding:"required"`
	Email string `json:"customerEmail" binding:"required"`
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

type createAccountResponseBody struct {
	Message   string `json:"message"`
	AccountID int    `json:"accountID"`
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

func (d database) createCustomerInDatabase(name string, email string) (int, error) {
	var customerID int
	// Insert data into the customers table and retrieve the inserted id
	err := d.DB.QueryRow("INSERT INTO customers (name, email) VALUES ($1, $2) RETURNING customer_id", name, email).Scan(&customerID)
	return customerID, err
}

func (d database) deleteCustomerFromDatabase(customerID int) error {
	// Delete data from the customers table
	_, err := d.DB.Exec("DELETE FROM customers WHERE customer_id = $1", customerID)
	return err
}

func (d database) getCustomerFromDatabase(customerID int) (string, string, error) {
	// Get customer data from the customers table
	var name, email string
	row := d.DB.QueryRow("SELECT name, email FROM customers WHERE customer_id = $1", customerID)
	err := row.Scan(&name, &email)
	if err != nil {
		return "", "", err
	}
	return name, email, nil
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

func sendCreateAccountRequest(customerID int, accountService string) (createAccountResponseBody, error) {
	payload := createAccountRequestPayload {
		CustomerID: customerID,
	}

	// Marshal the struct into a JSON-formatted byte slice
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return createAccountResponseBody{}, err
	}

	url := "http://" + accountService + "/createAccount"

	body, err := performPostRequest(&http.Client{}, url, jsonPayload)
	if err != nil {
		log.Printf("Error performing POST request: %v", err)
		return createAccountResponseBody{}, err
	}

	// Unmarshal the JSON response into a struct
	var response createAccountResponseBody
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Printf("Error unmarshaling JSON: %v", err)
		return createAccountResponseBody{}, err
	}

	return response, nil
}

func main() {
	gin.SetMode(gin.DebugMode)

	d := database{
		Host:     os.Getenv("CUSTOMER_SERVICE_DATABASE_HOST"),
		Port:     os.Getenv("CUSTOMER_SERVICE_DATABASE_PORT"),
		User:     os.Getenv("CUSTOMER_SERVICE_DATABASE_USER"),
		Password: os.Getenv("CUSTOMER_SERVICE_DATABASE_PASSWORD"),
		Name:     os.Getenv("CUSTOMER_SERVICE_DATABASE_NAME"),
		DB:       &sql.DB{},
	}

	err := d.init()
	if err != nil {
		return
	}

	accountService := os.Getenv("ACCOUNT_SERVICE_HOST_AND_PORT")

	// Create a new Gin router
	r := gin.Default()

	// Route for creating customers
	r.POST("/createCustomer", func(c *gin.Context) {
		var requestBody createRequestBody

		// Bind the JSON body to the RequestBody struct
		if err := c.BindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		customerID, err := d.createCustomerInDatabase(requestBody.Name, requestBody.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		response, err := sendCreateAccountRequest(customerID, accountService)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"message":    "success",
			"customerID": customerID,
			"accountID":  response.AccountID,
		})
	})

	// Route for deleting customers
	r.POST("/deleteCustomer", func(c *gin.Context) {
		var requestBody deleteRequestBody

		// Bind the JSON body to the RequestBody struct
		if err := c.BindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := d.deleteCustomerFromDatabase(requestBody.CustomerID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":    "success",
			"customerID": requestBody.CustomerID,
		})
	})

	// Route for retrieving customers data
	r.POST("/getCustomer", func(c *gin.Context) {
		var requestBody getRequestBody

		// Bind the JSON body to the RequestBody struct
		if err := c.BindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		name, email, err := d.getCustomerFromDatabase(requestBody.CustomerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":       "success",
			"customerID":    requestBody.CustomerID,
			"customerName":  name,
			"customerEmail": email,
		})
	})

	// Run the server on port 8080
	r.Run(":8080")
	defer d.DB.Close()
}
