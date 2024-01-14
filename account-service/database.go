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
