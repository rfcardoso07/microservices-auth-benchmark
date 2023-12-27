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

type createRequestBody struct {
	Name  string `json:"customerName" binding:"required"`
	Email string `json:"customerEmail" binding:"required"`
}

type deleteRequestBody struct {
	ID int `json:"customerID" binding:"required"`
}

type getRequestBody struct {
	ID int `json:"customerID" binding:"required"`
}

type database struct {
	host     string
	port     string
	user     string
	password string
	name     string
	db       *sql.DB
}

func (d *database) init() error {
	// Create connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		d.host, d.port, d.user, d.password, d.name)

	// Open a database connection and set up a connection pool
	var err error
	d.db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("Error opening database connection: %v", err)
		return err
	}

	// Set the maximum number of open (in-use + idle) connections
	d.db.SetMaxOpenConns(10)

	// Set the maximum number of idle connections in the pool
	d.db.SetMaxIdleConns(5)

	// Check if the database connection is alive
	err = d.db.Ping()
	if err != nil {
		log.Printf("Error pinging database: %v", err)
		return err
	}

	fmt.Println("Connected to the database")
	return nil
}

func (d database) createCustomerInDatabase(name string, email string) (int, error) {
	var id int
	// Insert data into the customers table and retrieve the inserted id
	err := d.db.QueryRow("INSERT INTO customers (name, email) VALUES ($1, $2) RETURNING id", name, email).Scan(&id)
	return id, err
}

func (d database) deleteCustomerFromDatabase(id int) error {
	// Delete data from the customers table
	_, err := d.db.Exec("DELETE FROM customers WHERE id = $1", id)
	return err
}

func (d database) getCustomerFromDatabase(id int) (string, string, error) {
	// Get customer data from the customers table
	var name, email string
	row := d.db.QueryRow("SELECT name, email FROM customers WHERE id = $1", id)
	err := row.Scan(&name, &email)
	if err != nil {
		return "", "", err
	}
	return name, email, nil
}

func main() {
	d := database{
		host:     os.Getenv("COSTUMER_SERVICE_DATABASE_HOST"),
		port:     os.Getenv("COSTUMER_SERVICE_DATABASE_PORT"),
		user:     os.Getenv("COSTUMER_SERVICE_DATABASE_USER"),
		password: os.Getenv("COSTUMER_SERVICE_DATABASE_PASSWORD"),
		name:     os.Getenv("COSTUMER_SERVICE_DATABASE_NAME"),
		db:       &sql.DB{},
	}

	err := d.init()
	if err != nil {
		return
	}

	// Create a new Gin router
	r := gin.Default()

	// Route for creating customers
	r.POST("/create", func(c *gin.Context) {
		var requestBody createRequestBody

		// Bind the JSON body to the RequestBody struct
		if err := c.BindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		id, err := d.createCustomerInDatabase(requestBody.Name, requestBody.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		if err := c.BindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := d.deleteCustomerFromDatabase(requestBody.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		if err := c.BindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		name, email, err := d.getCustomerFromDatabase(requestBody.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
	defer d.db.Close()
}
