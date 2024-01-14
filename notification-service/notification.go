package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"

	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(os.Getenv("GIN_MODE"))

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

	switch authPattern {
	case "NO_AUTH":
		// Route for sending notifications
		r.POST("/notify", func(c *gin.Context) {
			var requestBody notifyRequestBody

			// Bind the JSON body to the RequestBody struct
			if err := c.BindJSON(&requestBody); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			getAccountResponse, err := sendGetAccountRequest(requestBody.ReceiverID, accountService, "", "")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			getCustomerResponse, err := sendGetCustomerRequest(getAccountResponse.CustomerID, customerService, "", "")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			notificationID, err := d.registerNotificationInDatabase(requestBody.TransactionID, requestBody.ReceiverID, requestBody.Amount)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			log.Printf("Notifying %v of transaction %v", getCustomerResponse.CustomerEmail, requestBody.TransactionID)

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

	case "CENTRALIZED":
		// Route for sending notifications
		r.POST("/notify/:id/:password", func(c *gin.Context) {
			userID := c.Param("id")
			userPassword := c.Param("password")

			authResponse, err := sendAuthRequest(userID, userPassword, "WRITE", authService)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			if authResponse.AccessGranted {
				var requestBody notifyRequestBody

				// Bind the JSON body to the RequestBody struct
				if err := c.BindJSON(&requestBody); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				getAccountResponse, err := sendGetAccountRequest(requestBody.ReceiverID, accountService, userID, userPassword)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				getCustomerResponse, err := sendGetCustomerRequest(getAccountResponse.CustomerID, customerService, userID, userPassword)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				notificationID, err := d.registerNotificationInDatabase(requestBody.TransactionID, requestBody.ReceiverID, requestBody.Amount)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				log.Printf("Notifying %v of transaction %v", getCustomerResponse.CustomerEmail, requestBody.TransactionID)

				c.JSON(http.StatusOK, gin.H{
					"message":        "success",
					"notificationID": notificationID,
					"recipientEmail": getCustomerResponse.CustomerEmail,
				})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"message":       "accessDenied",
					"authenticated": authResponse.Authenticated,
					"authorized":    authResponse.Authorized,
				})
			}
		})

		// Route for retrieving notifications data
		r.POST("/getNotification/:id/:password", func(c *gin.Context) {
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
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"message":       "accessDenied",
					"authenticated": authResponse.Authenticated,
					"authorized":    authResponse.Authorized,
				})
			}
		})

	case "DECENTRALIZED":
		// Route for sending notifications
		r.POST("/notify/:id/:password", func(c *gin.Context) {
			userID := c.Param("id")
			userPassword := c.Param("password")

			authenticated, authorized, accessGranted, err := d.authenticateAndAuthorize(userID, userPassword, "WRITE")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			if accessGranted {
				var requestBody notifyRequestBody

				// Bind the JSON body to the RequestBody struct
				if err := c.BindJSON(&requestBody); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				getAccountResponse, err := sendGetAccountRequest(requestBody.ReceiverID, accountService, userID, userPassword)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				getCustomerResponse, err := sendGetCustomerRequest(getAccountResponse.CustomerID, customerService, userID, userPassword)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				notificationID, err := d.registerNotificationInDatabase(requestBody.TransactionID, requestBody.ReceiverID, requestBody.Amount)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				// This is where an e-mail would be sent.
				log.Printf("Notifying %v of transaction %v...", getCustomerResponse.CustomerEmail, requestBody.TransactionID)

				c.JSON(http.StatusOK, gin.H{
					"message":        "success",
					"notificationID": notificationID,
					"recipientEmail": getCustomerResponse.CustomerEmail,
				})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"message":       "accessDenied",
					"authenticated": authenticated,
					"authorized":    authorized,
				})
			}
		})

		// Route for retrieving notifications data
		r.POST("/getNotification/:id/:password", func(c *gin.Context) {
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

	// Run the server on port 8086
	r.Run(":8086")
	defer d.DB.Close()
}
