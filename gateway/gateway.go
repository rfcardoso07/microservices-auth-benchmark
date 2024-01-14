package main

import (
	"database/sql"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(os.Getenv("GIN_MODE"))

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

	router := gin.Default()

	customerServiceURL := os.Getenv("CUSTOMER_SERVICE_URL")
	accountServiceURL := os.Getenv("ACCOUNT_SERVICE_URL")
	transactionServiceURL := os.Getenv("TRANSACTION_SERVICE_URL")
	notificationServiceURL := os.Getenv("NOTIFICATION_SERVICE_URL")
	balanceServiceURL := os.Getenv("BALANCE_SERVICE_URL")

	gatewayAuth := os.Getenv("GATEWAY_AUTH")

	// Define the target URLs for different paths
	forwardURLs := map[string]string{
		"/createCustomer":           customerServiceURL,
		"/deleteCustomer":           customerServiceURL,
		"/getCustomer":              customerServiceURL,
		"/createAccount":            accountServiceURL,
		"/deleteAccount":            accountServiceURL,
		"/deleteAccountsByCustomer": accountServiceURL,
		"/getAccount":               accountServiceURL,
		"/getAccountsByCustomer":    accountServiceURL,
		"/addToBalance":             accountServiceURL,
		"/subtractFromBalance":      accountServiceURL,
		"/transferAmount":           transactionServiceURL,
		"/transferAmountAndNotify":  transactionServiceURL,
		"/getTransaction":           transactionServiceURL,
		"/notify":                   notificationServiceURL,
		"/getNotification":          notificationServiceURL,
		"/getBalanceByCustomer":     balanceServiceURL,
		"/getBalanceHistory":        balanceServiceURL,
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
		// Define a handler which enforces auth before forwarding requests to the appropriate services
		forwardHandler := func(c *gin.Context) {
			path := c.Request.URL.Path
			proxy, ok := proxies[path]
			if !ok {
				log.Printf("Could not find endpoint for path %s", path)
				c.JSON(http.StatusNotFound, gin.H{"error": "Endpoint not found"})
				return
			}

			// Extract userID and password from URL parameters
			userID := c.Param("id")
			userPassword := c.Param("password")

			authenticated, authorized, accessGranted, err := d.authenticateAndAuthorize(userID, userPassword, "WRITE")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			if accessGranted {
				// Forward the request to the corresponding target URL
				log.Printf("Received request at path %s, forwarding it to %s", path, forwardURLs[path])
				proxy.ServeHTTP(c.Writer, c.Request)
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"message":       "accessDenied",
					"authenticated": authenticated,
					"authorized":    authorized,
				})
			}
		}

		// Associate the forwardHandler with wildcard route
		router.POST("/:userID/:password", forwardHandler)

	} else {
		// Define a basic handler that forwards requests to the appropriate services
		forwardHandler := func(c *gin.Context) {
			path := c.Request.URL.Path
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

		// Associate the forwardHandler with wildcard route
		router.POST("/*path", forwardHandler)
	}

	// Run the Gin gateway on port 8000
	router.Run(":8000")
}
