package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func performPostRequest(client *http.Client, url string, payload []byte) ([]byte, error) {
	// Create a POST request with the JSON payload
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	// Set headers
	request.Header.Set("Content-Type", "application/json")

	// Send the request using the provided http.Client
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	statusCode := response.StatusCode
	if statusCode != http.StatusOK && statusCode != http.StatusUnauthorized {
		return nil, errors.New("Response has unexpected status code - " + strconv.Itoa(statusCode))
	}

	// Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func sendCreateAccountRequest(customerID int, accountService string, userID string, userPassword string) (createAccountResponseBody, error) {
	payload := createAccountRequestPayload{
		CustomerID: customerID,
	}

	// Marshal the struct into a JSON-formatted byte slice
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return createAccountResponseBody{}, err
	}

	url := "http://" + accountService + "/createAccount"
	if userID != "" && userPassword != "" {
		url = url + "/" + userID + "/" + userPassword
	}

	body, err := performPostRequest(&http.Client{}, url, jsonPayload)
	if err != nil {
		log.Printf("Error performing POST request: %v", err)
		return createAccountResponseBody{}, err
	}

	// Unmarshal the JSON response into a struct
	var response createAccountResponseBody
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Printf("Error unmarshaling JSON: %v", err)
		return createAccountResponseBody{}, err
	}

	if gin.IsDebugging() {
		fmt.Printf("%+v\n", response)
	}

	return response, nil
}

func sendDeleteAccountsByCustomerRequest(customerID int, accountService string, userID string, userPassword string) (deleteAccountsByCustomerResponseBody, error) {
	payload := deleteAccountsByCustomerRequestPayload{
		CustomerID: customerID,
	}

	// Marshal the struct into a JSON-formatted byte slice
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return deleteAccountsByCustomerResponseBody{}, err
	}

	url := "http://" + accountService + "/deleteAccountsByCustomer"
	if userID != "" && userPassword != "" {
		url = url + "/" + userID + "/" + userPassword
	}

	body, err := performPostRequest(&http.Client{}, url, jsonPayload)
	if err != nil {
		log.Printf("Error performing POST request: %v", err)
		return deleteAccountsByCustomerResponseBody{}, err
	}

	// Unmarshal the JSON response into a struct
	var response deleteAccountsByCustomerResponseBody
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Printf("Error unmarshaling JSON: %v", err)
		return deleteAccountsByCustomerResponseBody{}, err
	}

	if gin.IsDebugging() {
		fmt.Printf("%+v\n", response)
	}

	return response, nil
}

func sendAuthRequest(userID string, userPassword string, operation string, authService string) (authResponseBody, error) {
	payload := authRequestPayload{
		UserID:    userID,
		Password:  userPassword,
		Operation: operation,
	}

	// Marshal the struct into a JSON-formatted byte slice
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return authResponseBody{}, err
	}

	url := "http://" + authService + "/authenticateAndAuthorize"

	body, err := performPostRequest(&http.Client{}, url, jsonPayload)
	if err != nil {
		log.Printf("Error performing POST request: %v", err)
		return authResponseBody{}, err
	}

	// Unmarshal the JSON response into a struct
	var response authResponseBody
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Printf("Error unmarshaling JSON: %v", err)
		return authResponseBody{}, err
	}

	if gin.IsDebugging() {
		fmt.Printf("%+v\n", response)
	}

	return response, nil
}
