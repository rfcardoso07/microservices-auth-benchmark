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
	CustomerID int `json:"customerID" binding:"required"`
}

type deleteRequestBody struct {
	AccountID int `json:"accountID" binding:"required"`
}

type deleteByCustomerRequestBody struct {
	CustomerID int `json:"customerID" binding:"required"`
}

type getRequestBody struct {
	AccountID int `json:"accountID" binding:"required"`
}

type getByCustomerRequestBody struct {
	CustomerID int `json:"customerID" binding:"required"`
}

type addToBalanceRequestBody struct {
	AccountID int `json:"accountID" binding:"required"`
	Amount    int `json:"amount" binding:"required"`
}

type subtractFromBalanceRequestBody struct {
	AccountID int `json:"accountID" binding:"required"`
	Amount    int `json:"amount" binding:"required"`
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

func (d database) createAccountInDatabase(customerID int) (int, error) {
	var accountID int
	// Insert data into the accounts table and retrieve the inserted id
	err := d.DB.QueryRow("INSERT INTO accounts (customer_id, balance) VALUES ($1, $2) RETURNING account_id", customerID, 0).Scan(&accountID)
	return accountID, err
}

func (d database) deleteAccountFromDatabase(accountID int) error {
	// Delete data from the accounts table
	_, err := d.DB.Exec("DELETE FROM accounts WHERE account_id = $1", accountID)
	return err
}

func (d database) deleteAccountsFromDatabaseByCustomer(customerID int) ([]int, error) {
	// Delete data from the accounts table and return the deleted account IDs
	rows, err := d.DB.Query("DELETE FROM accounts WHERE customer_id = $1 RETURNING account_id", customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deletedAccountIDs []int

	// Iterate through the rows and add account IDs to the slice
	for rows.Next() {
		var deletedAccountID int
		if err := rows.Scan(&deletedAccountID); err != nil {
			return nil, err
		}
		deletedAccountIDs = append(deletedAccountIDs, deletedAccountID)
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return deletedAccountIDs, nil
}

func (d database) getAccountFromDatabase(accountID int) (int, int, error) {
	// Get account data from the accounts table
	var customerID, balance int
	row := d.DB.QueryRow("SELECT customer_id, balance FROM accounts WHERE account_id = $1", accountID)
	err := row.Scan(&customerID, &balance)
	if err != nil {
		return 0, 0, err
	}
	return customerID, balance, nil
}

func (d database) getAccountsFromDatabaseByCustomer(customerID int) ([]int, []int, error) {
	// Get account data from the accounts table
	var accountIDs, balances []int

	rows, err := d.DB.Query("SELECT account_id, balance FROM accounts WHERE customer_id = $1", customerID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	// Iterate through the rows and add values to slices
	for rows.Next() {
		var accountID, balance int
		if err := rows.Scan(&accountID, &balance); err != nil {
			return nil, nil, err
		}
		accountIDs = append(accountIDs, accountID)
		balances = append(balances, balance)
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return accountIDs, balances, nil
}


func (d database) addToAccountBalanceInDatabase(accountID int, amount int) error {
	// Add amount to account balance by updating accounts table
	_, err := d.DB.Exec("UPDATE accounts SET balance = balance + $1 WHERE account_id = $2", amount, accountID)
	return err
}

func (d database) subtractFromAccountBalanceInDatabase(accountID int, amount int) error {
	// Subtract amount from account balance by updating accounts table
	_, err := d.DB.Exec("UPDATE accounts SET balance = balance - $1 WHERE account_id = $2", amount, accountID)
	return err
}

func main() {
	gin.SetMode(gin.DebugMode)

	d := database{
		Host:     os.Getenv("ACCOUNT_SERVICE_DATABASE_HOST"),
		Port:     os.Getenv("ACCOUNT_SERVICE_DATABASE_PORT"),
		User:     os.Getenv("ACCOUNT_SERVICE_DATABASE_USER"),
		Password: os.Getenv("ACCOUNT_SERVICE_DATABASE_PASSWORD"),
		Name:     os.Getenv("ACCOUNT_SERVICE_DATABASE_NAME"),
		DB:       &sql.DB{},
	}

	err := d.init()
	if err != nil {
		return
	}

	// Create a new Gin router
	r := gin.Default()

	// Route for creating accounts
	r.POST("/createAccount", func(c *gin.Context) {
		var requestBody createRequestBody

		// Bind the JSON body to the RequestBody struct
		if err := c.BindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		accountID, err := d.createAccountInDatabase(requestBody.CustomerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"message":   "success",
			"accountID": accountID,
		})
	})

	// Route for deleting accounts
	r.POST("/deleteAccount", func(c *gin.Context) {
		var requestBody deleteRequestBody

		// Bind the JSON body to the RequestBody struct
		if err := c.BindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := d.deleteAccountFromDatabase(requestBody.AccountID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":   "success",
			"accountID": requestBody.AccountID,
		})
	})

	// Route for deleting accounts by customer
	r.POST("/deleteAccountsByCustomer", func(c *gin.Context) {
		var requestBody deleteByCustomerRequestBody

		// Bind the JSON body to the RequestBody struct
		if err := c.BindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		accountIDs, err := d.deleteAccountsFromDatabaseByCustomer(requestBody.CustomerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":   "success",
			"customerID": requestBody.CustomerID,
			"accountIDs": accountIDs,
		})
	})

	// Route for retrieving accounts data
	r.POST("/getAccount", func(c *gin.Context) {
		var requestBody getRequestBody

		// Bind the JSON body to the RequestBody struct
		if err := c.BindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		customerID, balance, err := d.getAccountFromDatabase(requestBody.AccountID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":    "success",
			"accountID":  requestBody.AccountID,
			"customerID": customerID,
			"balance":    balance,
		})
	})

	// Route for retrieving accounts data by customer
	r.POST("/getAccountsByCustomer", func(c *gin.Context) {
		var requestBody getByCustomerRequestBody

		// Bind the JSON body to the RequestBody struct
		if err := c.BindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		accountIDs, balances, err := d.getAccountsFromDatabaseByCustomer(requestBody.CustomerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":    "success",
			"customerID": requestBody.CustomerID,
			"accountIDs": accountIDs,
			"balances":   balances,
		})
	})

	// Route for adding amounts to account balances
	r.POST("/addToBalance", func(c *gin.Context) {
		var requestBody addToBalanceRequestBody

		// Bind the JSON body to the RequestBody struct
		if err := c.BindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := d.addToAccountBalanceInDatabase(requestBody.AccountID, requestBody.Amount); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":     "success",
			"accountID":   requestBody.AccountID,
			"amountAdded": requestBody.Amount,
		})
	})

	// Route for subtracting amounts from account balances
	r.POST("/subtractFromBalance", func(c *gin.Context) {
		var requestBody subtractFromBalanceRequestBody

		// Bind the JSON body to the RequestBody struct
		if err := c.BindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := d.subtractFromAccountBalanceInDatabase(requestBody.AccountID, requestBody.Amount); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":          "success",
			"accountID":        requestBody.AccountID,
			"amountSubtracted": requestBody.Amount,
		})
	})

	// Run the server on port 8082
	r.Run(":8082")
	defer d.DB.Close()
}
