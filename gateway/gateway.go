package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	customerServiceURL := os.Getenv("CUSTOMER_SERVICE_URL")
	accountServiceURL := os.Getenv("ACCOUNT_SERVICE_URL")
	transactionServiceURL := os.Getenv("TRANSACTION_SERVICE_URL")
	notificationServiceURL := os.Getenv("NOTIFICATION_SERVICE_URL")

	// Define the target URLs for different paths
	forwardURLs := map[string]string{
		"/createCustomer":           customerServiceURL,
		"/deleteCustomer":           customerServiceURL,
		"/getCustomer":              customerServiceURL,
		"/createAccount":            accountServiceURL,
		"/deleteAccount":            accountServiceURL,
		"/deleteAccountsByCustomer": accountServiceURL
		"/getAccount":               accountServiceURL,
		"/getAccountsByCustomer":    accountServiceURL
		"/addToBalance":             accountServiceURL,
		"/subtractFromBalance":      accountServiceURL,
		"/transferAmount":           transactionServiceURL,
		"/transferAmountAndNotify":  transactionServiceURL,
		"/getTransaction":           transactionServiceURL,
		"/notify":                   notificationServiceURL,
		"/getNotification":          notificationServiceURL,
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

	// Define a handler that selects the appropriate reverse proxy based on the request path
	forwardHandler := func(c *gin.Context) {
		path := c.Request.URL.Path
		proxy, ok := proxies[path]
		if !ok {
			log.Printf("Could not find endpoint for path %s", path)
			c.JSON(http.StatusNotFound, gin.H{"error": "Endpoint not found"})
			return
		}

		log.Printf("Received request at path %s e forwarded it to %s", path, proxies[path])

		// Forward the request to the corresponding target URL
		proxy.ServeHTTP(c.Writer, c.Request)
	}

	// Associate the forwardHandler with different paths
	for path, _ := range forwardURLs {
		router.Any(path, forwardHandler)
	}

	// Run the Gin server on port 8000
	router.Run(":8000")
}
