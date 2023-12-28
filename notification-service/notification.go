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

type notifyRequestBody struct {
	TransactionID  int    `json:"transactionID" binding:"required"`
	Amount         int    `json:"amount" binding:"required"`
	RecipientEmail string `json:"recipientEmail" binding:"required"`
}

type getRequestBody struct {
	NotificationID int `json:"notificationID" binding:"required"`
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

func (d database) registerNotificationInDatabase(transactionID int, amount int, recipientEmail string) (int, error) {
	var notificationID int
	// Insert data into the notifications table and retrieve the inserted id
	err := d.DB.QueryRow("INSERT INTO notifications (transaction_id, amount, recipient_email) VALUES ($1, $2, $3) RETURNING notification_id", transactionID, amount, recipientEmail).Scan(&notificationID)
	return notificationID, err
}

func (d database) getNotificationFromDatabase(notificationID int) (int, int, string, error) {
	// Get transaction data from the transactions table
	var transactionID, amount int
	var recipientEmail string
	row := d.DB.QueryRow("SELECT transaction_id, amount, recipient_email FROM notifications WHERE id = $1", notificationID)
	err := row.Scan(&transactionID, &amount, &recipientEmail)
	if err != nil {
		return 0, 0, "", err
	}
	return transactionID, amount, recipientEmail, nil
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

	// Route for sending notifications
	r.POST("/notify", func(c *gin.Context) {
		var requestBody notifyRequestBody

		// Bind the JSON body to the RequestBody struct
		if err := c.BindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		notificationID, err := d.registerNotificationInDatabase(requestBody.TransactionID, requestBody.Amount, requestBody.RecipientEmail)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		} else {
			log.Println("This is going to be an e-mail, but for now it is only a log message.")
			c.JSON(http.StatusOK, gin.H{
				"message":       "success",
				"transactionID": notificationID,
			})
		}
	})

	// Route for retrieving notifications data
	r.POST("/getNotification", func(c *gin.Context) {
		var requestBody getRequestBody

		// Bind the JSON body to the RequestBody struct
		if err := c.BindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		transactionID, amount, recipientEmail, err := d.getNotificationFromDatabase(requestBody.NotificationID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":        "success",
			"notificationID": requestBody.NotificationID,
			"transactionID":  transactionID,
			"amount":         amount,
			"recipientEmail": recipientEmail,
		})
	})

	// Run the server on port 8086
	r.Run(":8086")
	defer d.DB.Close()
}
