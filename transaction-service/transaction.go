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

type transferRequestBody struct {
	SenderID   int `json:"senderID" binding:"required"`
	ReceiverID int `json:"receiverID" binding:"required"`
	Amount     int `json:"amount" binding:"required"`
}

type getRequestBody struct {
	TransactionID int `json:"transactionID" binding:"required"`
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

	// Create a new Gin router
	r := gin.Default()

	// Route for creating accounts
	r.POST("/transferAmount", func(c *gin.Context) {
		var requestBody transferRequestBody

		// Bind the JSON body to the RequestBody struct
		if err := c.BindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		transactionID, err := d.createTransactionInDatabase(requestBody.SenderID, requestBody.ReceiverID, requestBody.Amount)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		} else {
			c.JSON(http.StatusOK, gin.H{
				"message":       "success",
				"transactionID": transactionID,
			})
		}
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
