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

	switch authPattern {
	case "NO_AUTH":
		// Route for performing transactions
		r.POST("/transferAmount", func(c *gin.Context) {
			var requestBody transferRequestBody

			// Bind the JSON body to the RequestBody struct
			if err := c.BindJSON(&requestBody); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			subtractResponse, err := sendSubtractFromAccountRequest(requestBody.SenderID, requestBody.Amount, accountService, "", "")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			addResponse, err := sendAddToAccountRequest(requestBody.ReceiverID, requestBody.Amount, accountService, "", "")
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

			subtractResponse, err := sendSubtractFromAccountRequest(requestBody.SenderID, requestBody.Amount, accountService, "", "")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			addResponse, err := sendAddToAccountRequest(requestBody.ReceiverID, requestBody.Amount, accountService, "", "")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			transactionID, err := d.createTransactionInDatabase(requestBody.SenderID, requestBody.ReceiverID, requestBody.Amount)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			notifyResponse, err := sendNotifyRequest(transactionID, requestBody.ReceiverID, requestBody.Amount, notificationService, "", "")
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

	case "CENTRALIZED":
		// Route for performing transactions
		r.POST("/transferAmount/:id/:password", func(c *gin.Context) {
			userID := c.Param("id")
			userPassword := c.Param("password")

			authResponse, err := sendAuthRequest(userID, userPassword, "WRITE", authService)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			if authResponse.AccessGranted {
				var requestBody transferRequestBody

				// Bind the JSON body to the RequestBody struct
				if err := c.BindJSON(&requestBody); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				subtractResponse, err := sendSubtractFromAccountRequest(requestBody.SenderID, requestBody.Amount, accountService, userID, userPassword)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				addResponse, err := sendAddToAccountRequest(requestBody.ReceiverID, requestBody.Amount, accountService, userID, userPassword)
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
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"message":       "accessDenied",
					"authenticated": authResponse.Authenticated,
					"authorized":    authResponse.Authorized,
				})
			}
		})

		// Route for performing transactions and notifying receivers
		r.POST("/transferAmountAndNotify/:id/:password", func(c *gin.Context) {
			userID := c.Param("id")
			userPassword := c.Param("password")

			authResponse, err := sendAuthRequest(userID, userPassword, "WRITE", authService)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			if authResponse.AccessGranted {
				var requestBody transferAndNotifyRequestBody

				// Bind the JSON body to the RequestBody struct
				if err := c.BindJSON(&requestBody); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				subtractResponse, err := sendSubtractFromAccountRequest(requestBody.SenderID, requestBody.Amount, accountService, userID, userPassword)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				addResponse, err := sendAddToAccountRequest(requestBody.ReceiverID, requestBody.Amount, accountService, userID, userPassword)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				transactionID, err := d.createTransactionInDatabase(requestBody.SenderID, requestBody.ReceiverID, requestBody.Amount)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				notifyResponse, err := sendNotifyRequest(transactionID, requestBody.ReceiverID, requestBody.Amount, notificationService, userID, userPassword)
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
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"message":       "accessDenied",
					"authenticated": authResponse.Authenticated,
					"authorized":    authResponse.Authorized,
				})
			}
		})

		// Route for retrieving transactions data
		r.POST("/getTransaction/:id/:password", func(c *gin.Context) {
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
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"message":       "accessDenied",
					"authenticated": authResponse.Authenticated,
					"authorized":    authResponse.Authorized,
				})
			}
		})

	case "DECENTRALIZED":
		// Route for performing transactions
		r.POST("/transferAmount/:id/:password", func(c *gin.Context) {
			userID := c.Param("id")
			userPassword := c.Param("password")

			authenticated, authorized, accessGranted, err := d.authenticateAndAuthorize(userID, userPassword, "WRITE")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			if accessGranted {
				var requestBody transferRequestBody

				// Bind the JSON body to the RequestBody struct
				if err := c.BindJSON(&requestBody); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				subtractResponse, err := sendSubtractFromAccountRequest(requestBody.SenderID, requestBody.Amount, accountService, userID, userPassword)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				addResponse, err := sendAddToAccountRequest(requestBody.ReceiverID, requestBody.Amount, accountService, userID, userPassword)
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
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"message":       "accessDenied",
					"authenticated": authenticated,
					"authorized":    authorized,
				})
			}
		})

		// Route for performing transactions and notifying receivers
		r.POST("/transferAmountAndNotify/:id/:password", func(c *gin.Context) {
			userID := c.Param("id")
			userPassword := c.Param("password")

			authenticated, authorized, accessGranted, err := d.authenticateAndAuthorize(userID, userPassword, "WRITE")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			if accessGranted {
				var requestBody transferAndNotifyRequestBody

				// Bind the JSON body to the RequestBody struct
				if err := c.BindJSON(&requestBody); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				subtractResponse, err := sendSubtractFromAccountRequest(requestBody.SenderID, requestBody.Amount, accountService, userID, userPassword)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				addResponse, err := sendAddToAccountRequest(requestBody.ReceiverID, requestBody.Amount, accountService, userID, userPassword)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				transactionID, err := d.createTransactionInDatabase(requestBody.SenderID, requestBody.ReceiverID, requestBody.Amount)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				notifyResponse, err := sendNotifyRequest(transactionID, requestBody.ReceiverID, requestBody.Amount, notificationService, userID, userPassword)
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
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"message":       "accessDenied",
					"authenticated": authenticated,
					"authorized":    authorized,
				})
			}
		})

		// Route for retrieving transactions data
		r.POST("/getTransaction/:id/:password", func(c *gin.Context) {
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

	// Run the server on port 8084
	r.Run(":8084")
	defer d.DB.Close()
}
