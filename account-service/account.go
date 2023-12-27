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
	customerID int `json:"customerID" binding:"required"`
}

type deleteRequestBody struct {
	accountID int `json:"accountID" binding:"required"`
}

type database struct {
	host     string
	port     string
	user     string
	password string
	name     string
	db       *sql.DB
}

func (d *database) init() error {
	// Create connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		d.host, d.port, d.user, d.password, d.name)

	// Open a database connection and set up a connection pool
	var err error
	d.db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("Error opening database connection: %v", err)
		return err
	}

	// Set the maximum number of open (in-use + idle) connections
	d.db.SetMaxOpenConns(10)

	// Set the maximum number of idle connections in the pool
	d.db.SetMaxIdleConns(5)

	// Check if the database connection is alive
	err = d.db.Ping()
	if err != nil {
		log.Printf("Error pinging database: %v", err)
		return err
	}

	fmt.Println("Connected to the database")
	return nil
}

func (d database) createAccountInDatabase(customerID int) (int, error) {
	var accountID int
	// Insert data into the accounts table and retrieve the inserted id
	err := d.db.QueryRow("INSERT INTO accounts (customer_id, balance) VALUES ($1, $2) RETURNING account_id", customerID, 0).Scan(&accountID)
	return accountID, err
}

func (d database) deleteAccountFromDatabase(accountID int) error {
	// Delete data from the accounts table
	_, err := d.db.Exec("DELETE FROM accounts WHERE account_id = $1", accountID)
	return err
}

func main() {
	gin.SetMode(gin.DebugMode)

	d := database{
		host:     os.Getenv("ACCOUNT_SERVICE_DATABASE_HOST"),
		port:     os.Getenv("ACCOUNT_SERVICE_DATABASE_PORT"),
		user:     os.Getenv("ACCOUNT_SERVICE_DATABASE_USER"),
		password: os.Getenv("ACCOUNT_SERVICE_DATABASE_PASSWORD"),
		name:     os.Getenv("ACCOUNT_SERVICE_DATABASE_NAME"),
		db:       &sql.DB{},
	}

	err := d.init()
	if err != nil {
		return
	}

	// Create a new Gin router
	r := gin.Default()

	// Route for creating accounts
	r.POST("/create", func(c *gin.Context) {
		var requestBody createRequestBody

		// Bind the JSON body to the RequestBody struct
		if err := c.ShouldBindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		accountID, err := d.createAccountInDatabase(requestBody.customerID)
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
	r.POST("/delete", func(c *gin.Context) {
		var requestBody deleteRequestBody

		// Bind the JSON body to the RequestBody struct
		if err := c.ShouldBindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := d.deleteAccountFromDatabase(requestBody.accountID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":   "success",
			"accountID": requestBody.accountID,
		})
	})

	// Run the server on port 8000
	r.Run(":8001")
	defer d.db.Close()
}
