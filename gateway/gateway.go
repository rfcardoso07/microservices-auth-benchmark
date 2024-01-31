package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	_ "github.com/lib/pq"

	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(os.Getenv("GIN_MODE"))

	router := gin.Default()

	customerServiceURL := os.Getenv("CUSTOMER_SERVICE_URL")
	accountServiceURL := os.Getenv("ACCOUNT_SERVICE_URL")
	transactionServiceURL := os.Getenv("TRANSACTION_SERVICE_URL")
	notificationServiceURL := os.Getenv("NOTIFICATION_SERVICE_URL")
	balanceServiceURL := os.Getenv("BALANCE_SERVICE_URL")

	gatewayAuth := os.Getenv("GATEWAY_AUTH")

	// Define the target URLs for different paths
	forwardURLs := map[string]string{
		"createCustomer":           customerServiceURL,
		"deleteCustomer":           customerServiceURL,
		"getCustomer":              customerServiceURL,
		"createAccount":            accountServiceURL,
		"deleteAccount":            accountServiceURL,
		"deleteAccountsByCustomer": accountServiceURL,
		"getAccount":               accountServiceURL,
		"getAccountsByCustomer":    accountServiceURL,
		"addToBalance":             accountServiceURL,
		"subtractFromBalance":      accountServiceURL,
		"transferAmount":           transactionServiceURL,
		"transferAmountAndNotify":  transactionServiceURL,
		"getTransaction":           transactionServiceURL,
		"notify":                   notificationServiceURL,
		"getNotification":          notificationServiceURL,
		"getBalanceByCustomer":     balanceServiceURL,
		"getBalanceHistory":        balanceServiceURL,
	}

	// Define the operation types for different paths
	operationTypes := map[string]string{
		"createCustomer":           "WRITE",
		"deleteCustomer":           "DELETE",
		"getCustomer":              "READ",
		"createAccount":            "WRITE",
		"deleteAccount":            "DELETE",
		"deleteAccountsByCustomer": "DELETE",
		"getAccount":               "READ",
		"getAccountsByCustomer":    "READ",
		"addToBalance":             "WRITE",
		"subtractFromBalance":      "WRITE",
		"transferAmount":           "WRITE",
		"transferAmountAndNotify":  "WRITE",
		"getTransaction":           "READ",
		"notify":                   "WRITE",
		"getNotification":          "READ",
		"getBalanceByCustomer":     "READ",
		"getBalanceHistory":        "READ",
	}

	// Create reverse proxies for each target URL
	proxies := make(map[string]*httputil.ReverseProxy)
	for path, targetURL := range forwardURLs {
		target, err := url.Parse(targetURL)
		if err != nil {
			log.Printf("Error parsing target URL: %v", err)
		}
		proxies[path] = httputil.NewSingleHostReverseProxy(target)
	}

	if gatewayAuth == "TRUE" {
		d := database{
			Host:     os.Getenv("GATEWAY_DATABASE_HOST"),
			Port:     os.Getenv("GATEWAY_DATABASE_PORT"),
			User:     os.Getenv("GATEWAY_DATABASE_USER"),
			Password: os.Getenv("GATEWAY_DATABASE_PASSWORD"),
			Name:     os.Getenv("GATEWAY_DATABASE_NAME"),
			DB:       &sql.DB{},
		}

		err := d.init()
		if err != nil {
			return
		}

		defer d.DB.Close()

		// Define a handler which enforces auth before forwarding requests to the appropriate services
		authHandler := func(c *gin.Context) {
			path := c.Param("path")
			proxy, ok := proxies[path]
			if !ok {
				log.Printf("Could not find endpoint for path %s", path)
				c.JSON(http.StatusNotFound, gin.H{"error": "Endpoint not found"})
				return
			}

			// Extract userID and password from URL parameters
			userID := c.Param("userID")
			userPassword := c.Param("password")

			authenticated, authorized, accessGranted, err := d.authenticateAndAuthorize(userID, userPassword, operationTypes[path])
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			if accessGranted {
				// Forward the request to the corresponding target URL
				log.Printf("Received request at path %s, forwarding it to %s", path, forwardURLs[path])
				// Trim user ID and password from the request
				c.Request.URL.Path = strings.TrimSuffix(c.Request.URL.Path, fmt.Sprintf("/%s/%s", userID, userPassword))
				proxy.ServeHTTP(c.Writer, c.Request)
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"message":       "accessDenied",
					"authenticated": authenticated,
					"authorized":    authorized,
				})
			}
		}

		// Associate auth handler with wildcard route
		router.POST("/:path/:userID/:password", authHandler)

	} else {
		// Define a basic handler that forwards requests to the appropriate services
		forwardHandler := func(c *gin.Context) {
			path := c.Param("path")
			proxy, ok := proxies[path]
			if !ok {
				log.Printf("Could not find endpoint for path %s", path)
				c.JSON(http.StatusNotFound, gin.H{"error": "Endpoint not found"})
				return
			}

			// Forward the request to the corresponding target URL
			log.Printf("Received request at path %s, forwarding it to %s", path, forwardURLs[path])
			proxy.ServeHTTP(c.Writer, c.Request)
		}

		// Associate forward handler with wildcard routes
		router.POST("/:path", forwardHandler)
		router.POST("/:path/:userID/:password", forwardHandler)
	}

	// Run the Gin gateway on port 8000
	router.Run(":8000")
}
