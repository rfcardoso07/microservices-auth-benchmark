package main

import (
	"database/sql"
	"net/http"
	"os"

	_ "github.com/lib/pq"

	"github.com/gin-gonic/gin"
)

func hasPermission(operation string, permissions userPermissions) bool {
	switch operation {
	case "READ":
		if permissions.CanRead {
			return true
		}
		return false
	case "WRITE":
		if permissions.CanWrite {
			return true
		}
		return false
	case "DELETE":
		if permissions.CanDelete {
			return true
		}
		return false
	default:
		return false
	}
}

func main() {
	gin.SetMode(gin.DebugMode)

	d := database{
		Host:     os.Getenv("AUTH_SERVICE_DATABASE_HOST"),
		Port:     os.Getenv("AUTH_SERVICE_DATABASE_PORT"),
		User:     os.Getenv("AUTH_SERVICE_DATABASE_USER"),
		Password: os.Getenv("AUTH_SERVICE_DATABASE_PASSWORD"),
		Name:     os.Getenv("AUTH_SERVICE_DATABASE_NAME"),
		DB:       &sql.DB{},
	}

	err := d.init()
	if err != nil {
		return
	}

	// Create a new Gin router
	r := gin.Default()

	// Route for creating customers
	r.POST("/authenticateAndAuthorize", func(c *gin.Context) {
		var requestBody authRequestBody

		// Bind the JSON body to the RequestBody struct
		if err := c.BindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		authorized := false
		accessGranted := false

		authenticated, permissions, err := d.searchForUserInDatabase(requestBody.UserID, requestBody.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if authenticated {
			authorized = hasPermission(requestBody.Operation, permissions)
			if authorized {
				accessGranted = true
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message":        "success",
			"authenticated":  authenticated,
			"authorized":     authorized,
			"accessGranted:": accessGranted,
		})
	})

	// Run the server on port 8090
	r.Run(":8090")
	defer d.DB.Close()
}
