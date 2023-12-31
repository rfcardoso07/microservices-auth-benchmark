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

type authRequestBody struct {
	UserID    string `json:"userID" binding:"required"`
	Password  string `json:"password" binding:"required"`
	Operation string `json:"operation" binding:"required"`
}

type userPermissions struct {
	CanRead   bool
	CanWrite  bool
	CanDelete bool
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

func (d database) searchForUserInDatabase(userID string, password string) (bool, userPermissions, error) {
	// Search for userID and password in the users table and retrieve permissions
	var permissions userPermissions
	row := d.DB.QueryRow("SELECT can_read, can_write, can_delete FROM users WHERE user_id = $1 AND user_password = $2", userID, password)
	err := row.Scan(&permissions.CanRead, &permissions.CanWrite, &permissions.CanDelete)

	if err != nil {
		if err == sql.ErrNoRows {
			// Not actually an error, just means there was no match (user + password)
			return false, userPermissions{}, nil
		} else {
			return false, userPermissions{}, err
		}
	}

	return true, permissions, nil
}

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
