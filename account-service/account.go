package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"

	"github.com/gin-gonic/gin"
)

type createRequestBody struct {
	CustomerID int `json:"customerID" binding:"required"`
}

type deleteRequestBody struct {
	AccountID int `json:"accountID" binding:"required"`
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

func (d database) createAccountInDatabase(customerID int) (int, error) {
	var accountID int
	// Insert data into the accounts table and retrieve the inserted id
	err := d.DB.QueryRow("INSERT INTO accounts (customer_id, balance) VALUES ($1, $2) RETURNING account_id", customerID, 0).Scan(&accountID)
	return accountID, err
}

func (d database) deleteAccountFromDatabase(accountID int) error {
	// Delete data from the accounts table
	_, err := d.DB.Exec("DELETE FROM accounts WHERE account_id = $1", accountID)
	return err
}

func main() {
	gin.SetMode(gin.DebugMode)

	d := database{
		Host:     os.Getenv("ACCOUNT_SERVICE_DATABASE_HOST"),
		Port:     os.Getenv("ACCOUNT_SERVICE_DATABASE_PORT"),
		User:     os.Getenv("ACCOUNT_SERVICE_DATABASE_USER"),
		Password: os.Getenv("ACCOUNT_SERVICE_DATABASE_PASSWORD"),
		Name:     os.Getenv("ACCOUNT_SERVICE_DATABASE_NAME"),
		DB:       &sql.DB{},
	}

	err := d.init()
	if err != nil {
		return
	}

	// Create a new Gin router
	r := gin.Default()

	// Route for creating accounts
	r.POST("/createAccount", func(c *gin.Context) {
		var requestBody createRequestBody

		// Bind the JSON body to the RequestBody struct
		if err := c.BindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		accountID, err := d.createAccountInDatabase(requestBody.CustomerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		} else {
			c.JSON(http.StatusOK, gin.H{
				"message":   "success",
				"accountID": accountID,
			})
		}
	})

	// Route for deleting accounts
	r.POST("/deleteAccount", func(c *gin.Context) {
		var requestBody deleteRequestBody

		// Bind the JSON body to the RequestBody struct
		if err := c.BindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := d.deleteAccountFromDatabase(requestBody.AccountID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":   "success",
			"accountID": requestBody.AccountID,
		})
	})

	// Run the server on port 8082
	r.Run(":8082")
	defer d.db.Close()
}
