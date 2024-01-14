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
