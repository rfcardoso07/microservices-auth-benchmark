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

func sendAddToAccountRequest(accountID int, amount int, accountService string, userID string, userPassword string) (addToAccountResponseBody, error) {
	payload := addToAccountRequestPayload{
		AccountID: accountID,
		Amount:    amount,
	}

	// Marshal the struct into a JSON-formatted byte slice
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return addToAccountResponseBody{}, err
	}

	url := "http://" + accountService + "/addToBalance"
	if userID != "" && userPassword != "" {
		url = url + "/" + userID + "/" + userPassword
	}

	body, err := performPostRequest(&http.Client{}, url, jsonPayload)
	if err != nil {
		log.Printf("Error performing POST request: %v", err)
		return addToAccountResponseBody{}, err
	}

	// Unmarshal the JSON response into a struct
	var response addToAccountResponseBody
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Printf("Error unmarshaling JSON: %v", err)
		return addToAccountResponseBody{}, err
	}

	if gin.IsDebugging() {
		fmt.Println(string(body))
	}

	return response, nil
}

func sendSubtractFromAccountRequest(accountID int, amount int, accountService string, userID string, userPassword string) (subtractFromAccountResponseBody, error) {
	payload := subtractFromAccountRequestPayload{
		AccountID: accountID,
		Amount:    amount,
	}

	// Marshal the struct into a JSON-formatted byte slice
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return subtractFromAccountResponseBody{}, err
	}

	url := "http://" + accountService + "/subtractFromBalance"
	if userID != "" && userPassword != "" {
		url = url + "/" + userID + "/" + userPassword
	}

	body, err := performPostRequest(&http.Client{}, url, jsonPayload)
	if err != nil {
		log.Printf("Error performing POST request: %v", err)
		return subtractFromAccountResponseBody{}, err
	}

	// Unmarshal the JSON response into a struct
	var response subtractFromAccountResponseBody
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Printf("Error unmarshaling JSON: %v", err)
		return subtractFromAccountResponseBody{}, err
	}

	if gin.IsDebugging() {
		fmt.Println(string(body))
	}

	return response, nil
}

func sendNotifyRequest(transactionID int, receiverID int, amount int, notificationService string, userID string, userPassword string) (notifyResponseBody, error) {
	payload := notifyRequestPayload{
		TransactionID: transactionID,
		ReceiverID:    receiverID,
		Amount:        amount,
	}

	// Marshal the struct into a JSON-formatted byte slice
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return notifyResponseBody{}, err
	}

	url := "http://" + notificationService + "/notify"
	if userID != "" && userPassword != "" {
		url = url + "/" + userID + "/" + userPassword
	}

	body, err := performPostRequest(&http.Client{}, url, jsonPayload)
	if err != nil {
		log.Printf("Error performing POST request: %v", err)
		return notifyResponseBody{}, err
	}

	// Unmarshal the JSON response into a struct
	var response notifyResponseBody
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Printf("Error unmarshaling JSON: %v", err)
		return notifyResponseBody{}, err
	}

	if gin.IsDebugging() {
		fmt.Println(string(body))
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
		fmt.Println(string(body))
	}

	return response, nil
}
