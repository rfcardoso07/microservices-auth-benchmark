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

	defer d.DB.Close()

	accountService := os.Getenv("ACCOUNT_SERVICE_HOST_AND_PORT")
	authService := os.Getenv("AUTH_SERVICE_HOST_AND_PORT")
	authPattern := os.Getenv("APPLICATION_AUTH_PATTERN")

	// Create a new Gin router
	r := gin.Default()

	switch authPattern {
	case "NO_AUTH":
		// Route for creating customers
		r.POST("/createCustomer", func(c *gin.Context) {
			var requestBody createRequestBody

			// Bind the JSON body to the RequestBody struct
			if err := c.BindJSON(&requestBody); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			customerID, err := d.createCustomerInDatabase(requestBody.CustomerName, requestBody.CustomerEmail)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			response, err := sendCreateAccountRequest(customerID, accountService, "", "")
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

			response, err := sendDeleteAccountsByCustomerRequest(requestBody.CustomerID, accountService, "", "")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			if err := d.deleteCustomerFromDatabase(requestBody.CustomerID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"message":    "success",
				"customerID": requestBody.CustomerID,
				"accountIDs": response.AccountIDs,
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

	case "CENTRALIZED":
		// Route for creating customers
		r.POST("/createCustomer/:id/:password", func(c *gin.Context) {
			userID := c.Param("id")
			userPassword := c.Param("password")

			authResponse, err := sendAuthRequest(userID, userPassword, "WRITE", authService)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			if authResponse.AccessGranted {
				var requestBody createRequestBody

				// Bind the JSON body to the RequestBody struct
				if err := c.BindJSON(&requestBody); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				customerID, err := d.createCustomerInDatabase(requestBody.CustomerName, requestBody.CustomerEmail)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				response, err := sendCreateAccountRequest(customerID, accountService, userID, userPassword)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"message":    "success",
					"customerID": customerID,
					"accountID":  response.AccountID,
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
		r.POST("/deleteCustomer/:id/:password", func(c *gin.Context) {
			userID := c.Param("id")
			userPassword := c.Param("password")

			authResponse, err := sendAuthRequest(userID, userPassword, "DELETE", authService)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			if authResponse.AccessGranted {
				var requestBody deleteRequestBody

				// Bind the JSON body to the RequestBody struct
				if err := c.BindJSON(&requestBody); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				response, err := sendDeleteAccountsByCustomerRequest(requestBody.CustomerID, accountService, userID, userPassword)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				if err := d.deleteCustomerFromDatabase(requestBody.CustomerID); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"message":    "success",
					"customerID": requestBody.CustomerID,
					"accountIDs": response.AccountIDs,
				})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"message":       "accessDenied",
					"authenticated": authResponse.Authenticated,
					"authorized":    authResponse.Authorized,
				})
			}
		})

		// Route for retrieving customers data
		r.POST("/getCustomer/:id/:password", func(c *gin.Context) {
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
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"message":       "accessDenied",
					"authenticated": authResponse.Authenticated,
					"authorized":    authResponse.Authorized,
				})
			}
		})

	case "DECENTRALIZED":
		r.POST("/createCustomer/:id/:password", func(c *gin.Context) {
			userID := c.Param("id")
			userPassword := c.Param("password")

			authenticated, authorized, accessGranted, err := d.authenticateAndAuthorize(userID, userPassword, "WRITE")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			if accessGranted {
				var requestBody createRequestBody

				// Bind the JSON body to the RequestBody struct
				if err := c.BindJSON(&requestBody); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				customerID, err := d.createCustomerInDatabase(requestBody.CustomerName, requestBody.CustomerEmail)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				response, err := sendCreateAccountRequest(customerID, accountService, userID, userPassword)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"message":    "success",
					"customerID": customerID,
					"accountID":  response.AccountID,
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
		r.POST("/deleteCustomer/:id/:password", func(c *gin.Context) {
			userID := c.Param("id")
			userPassword := c.Param("password")

			authenticated, authorized, accessGranted, err := d.authenticateAndAuthorize(userID, userPassword, "DELETE")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			if accessGranted {
				var requestBody deleteRequestBody

				// Bind the JSON body to the RequestBody struct
				if err := c.BindJSON(&requestBody); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				response, err := sendDeleteAccountsByCustomerRequest(requestBody.CustomerID, accountService, userID, userPassword)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				if err := d.deleteCustomerFromDatabase(requestBody.CustomerID); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"message":    "success",
					"customerID": requestBody.CustomerID,
					"accountIDs": response.AccountIDs,
				})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"message":       "accessDenied",
					"authenticated": authenticated,
					"authorized":    authorized,
				})
			}
		})

		// Route for retrieving customers data
		r.POST("/getCustomer/:id/:password", func(c *gin.Context) {
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

	// Run the server on port 8080
	r.Run(":8080")
}
