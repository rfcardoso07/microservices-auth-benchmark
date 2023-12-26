package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/lib/pq"

	"github.com/gin-gonic/gin"
)

type createRequestBody struct {
	Name  string `json:"customerName"`
	Email string `json:"customerEmail"`
}

type deleteRequestBody struct {
	ID string `json:"customerID"`
}

type getRequestBody struct {
	ID string `json:"customerID"`
}

func initDB() (*sql.DB, error) {
	var db *sql.DB

	host := os.Getenv("COSTUMER_SERVICE_DATABASE_HOST")
	port := os.Getenv("COSTUMER_SERVICE_DATABASE_PORT")
	user := os.Getenv("COSTUMER_SERVICE_DATABASE_USER")
	password := os.Getenv("COSTUMER_SERVICE_DATABASE_PASSWORD")
	dbname := os.Getenv("COSTUMER_SERVICE_DATABASE_NAME")

	// Create connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	fmt.Println(connStr)

	// Open a database connection and set up a connection pool
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("Error opening database connection: %v", err)
		return nil, err
	}

	// Set the maximum number of open (in-use + idle) connections
	db.SetMaxOpenConns(10)

	// Set the maximum number of idle connections in the pool
	db.SetMaxIdleConns(5)

	// Check if the database connection is alive
	err = db.Ping()
	if err != nil {
		log.Printf("Error pinging database: %v", err)
		return nil, err
	}

	fmt.Println("Connected to the database")
	return db, nil
}

func createCustomerInDatabase(db *sql.DB, name string, email string) (string, error) {
	var id int
	// Insert data into the customers table and retrieve the inserted id
	err := db.QueryRow("INSERT INTO customers (name, email) VALUES ($1, $2) RETURNING id", name, email).Scan(&id)
	return strconv.Itoa(id), err
}

func deleteCustomerFromDatabase(db *sql.DB, id string) error {
	// Delete data from the customers table
	_, err := db.Exec("DELETE FROM customers WHERE id = $1", id)
	return err
}

func getCustomerFromDatabase(db *sql.DB, id string) (string, string, error) {
	// Get customer data from the customers table
	var name, email string
	row := db.QueryRow("SELECT name, email FROM customers WHERE id = $1", id)
	err := row.Scan(&name, &email)
	if err != nil {
		return "", "", err
	}
	return name, email, nil
}

func main() {
	db, err := initDB()
	if err != nil {
		return
	}

	// Create a new Gin router
	r := gin.Default()

	// Route for creating customers
	r.POST("/create", func(c *gin.Context) {
		var requestBody createRequestBody

		// Bind the JSON body to the RequestBody struct
		if err := c.ShouldBindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		id, err := createCustomerInDatabase(db, requestBody.Name, requestBody.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		} else {
			c.JSON(http.StatusOK, gin.H{
				"message":    "success",
				"customerID": id,
			})
		}
	})

	// Route for deleting customers
	r.POST("/delete", func(c *gin.Context) {
		var requestBody deleteRequestBody

		// Bind the JSON body to the RequestBody struct
		if err := c.ShouldBindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := deleteCustomerFromDatabase(db, requestBody.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":    "success",
			"customerID": requestBody.ID,
		})
	})

	// Route for retrieving customers data
	r.GET("/get", func(c *gin.Context) {
		var requestBody getRequestBody

		// Bind the JSON body to the RequestBody struct
		if err := c.ShouldBindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		name, email, err := getCustomerFromDatabase(db, requestBody.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":       "success",
			"customerID":    requestBody.ID,
			"customerName":  name,
			"customerEmail": email,
		})
	})

	// Run the server on port 8000
	r.Run(":8000")
	defer db.Close()
}
