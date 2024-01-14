package main

import (
	"database/sql"
	"fmt"
	"log"
)

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

func (d database) createCustomerInDatabase(name string, email string) (int, error) {
	var customerID int
	// Insert data into the customers table and retrieve the inserted id
	err := d.DB.QueryRow("INSERT INTO customers (name, email) VALUES ($1, $2) RETURNING customer_id", name, email).Scan(&customerID)
	return customerID, err
}

func (d database) deleteCustomerFromDatabase(customerID int) error {
	// Delete data from the customers table
	_, err := d.DB.Exec("DELETE FROM customers WHERE customer_id = $1", customerID)
	return err
}

func (d database) getCustomerFromDatabase(customerID int) (string, string, error) {
	// Get customer data from the customers table
	var name, email string
	row := d.DB.QueryRow("SELECT name, email FROM customers WHERE customer_id = $1", customerID)
	err := row.Scan(&name, &email)
	if err != nil {
		return "", "", err
	}
	return name, email, nil
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

func (d database) authenticateAndAuthorize(userID string, userPassword string, operation string) (bool, bool, bool, error) {
	authorized := false
	accessGranted := false

	authenticated, permissions, err := d.searchForUserInDatabase(userID, userPassword)
	if err != nil {
		return false, false, false, err
	}

	if authenticated {
		authorized = hasPermission(operation, permissions)
		if authorized {
			accessGranted = true
		}
	}

	return authenticated, authorized, accessGranted, nil
}
