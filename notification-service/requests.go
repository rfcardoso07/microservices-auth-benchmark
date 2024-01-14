package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
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

	// Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func sendGetAccountRequest(accountID int, accountService string, userID string, userPassword string) (getAccountResponseBody, error) {
	payload := getAccountRequestPayload{
		AccountID: accountID,
	}

	// Marshal the struct into a JSON-formatted byte slice
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return getAccountResponseBody{}, err
	}

	url := "http://" + accountService + "/getAccount"
	if userID != "" && userPassword != "" {
		url = url + "/" + userID + "/" + userPassword
	}

	body, err := performPostRequest(&http.Client{}, url, jsonPayload)
	if err != nil {
		log.Printf("Error performing POST request: %v", err)
		return getAccountResponseBody{}, err
	}

	// Unmarshal the JSON response into a struct
	var response getAccountResponseBody
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Printf("Error unmarshaling JSON: %v", err)
		return getAccountResponseBody{}, err
	}

	return response, nil
}

func sendGetCustomerRequest(customerID int, customerService string, userID string, userPassword string) (getCustomerResponseBody, error) {
	payload := getCustomerRequestPayload{
		CustomerID: customerID,
	}

	// Marshal the struct into a JSON-formatted byte slice
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return getCustomerResponseBody{}, err
	}

	url := "http://" + customerService + "/getCustomer"
	if userID != "" && userPassword != "" {
		url = url + "/" + userID + "/" + userPassword
	}

	body, err := performPostRequest(&http.Client{}, url, jsonPayload)
	if err != nil {
		log.Printf("Error performing POST request: %v", err)
		return getCustomerResponseBody{}, err
	}

	// Unmarshal the JSON response into a struct
	var response getCustomerResponseBody
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Printf("Error unmarshaling JSON: %v", err)
		return getCustomerResponseBody{}, err
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

	return response, nil
}
