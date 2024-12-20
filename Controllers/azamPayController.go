package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type AuthResponse struct {
	Data struct {
		AccessToken string    `json:"accessToken"`
		Expire      time.Time `json:"expire"`
	} `json:"data"`
	Message    string `json:"message"`
	Success    bool   `json:"success"`
	StatusCode int    `json:"statusCode"`
}

type MNOCheckoutResponse struct {
	TransactionId string `json:"transactionId"`
	Message       string `json:"message"`
	Success       bool   `json:"success"`
}

func AuthToken() string {
	authUrl := os.Getenv("AZAMPAY_AUTH")
	appName := os.Getenv("AZAMPAY_APP_NAME")
	clientID := os.Getenv("AZAMPAY_CLIENT_ID")
	clientSecret := os.Getenv("AZAMPAY_CLIENT_SECRET")

	// Create JSON body
	jsonBody := fmt.Sprintf(`{
		"appName": "%s",
		"clientId": "%s",
		"clientSecret": "%s"
	}`, appName, clientID, clientSecret)

	bodyReader := bytes.NewReader([]byte(jsonBody))

	// Create HTTP request
	req, err := http.NewRequest(http.MethodPost, authUrl, bodyReader)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)

	}

	req.Header.Set("Content-Type", "application/json")

	// Send HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Authentication failed with status code: %d", resp.StatusCode)
	}

	// Parse the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	// Unmarshal the response into the struct
	var authResp AuthResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		log.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Return the parsed auth token
	return authResp.Data.AccessToken
}

func MNOCheckout(accountNumber, amount, provider string) (bool, string) {
	// Generate the token
	token := AuthToken()
	externalID := "h87yh8hu874r98U9J98U9T6TGf562gd323rrf"

	// Create JSON body
	jsonBody := fmt.Sprintf(`{
		"accountNumber": "%s",
		"additionalProperties":{
			"property1": null,
			"property2": null
		},
		"amount": "%s",
		"currency": "TZS",
		"externalId":"%s",
		"provider": "%s"
	}`, accountNumber, amount, externalID, provider)

	checkoutUrl := os.Getenv("AZAMPAY_CHECKOUT")
	bodyReader := bytes.NewReader([]byte(jsonBody))

	// Create HTTP request
	req, err := http.NewRequest(http.MethodPost, checkoutUrl, bodyReader)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}

	// Set headers separately
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// fmt.Println(req.Header)

	// Send HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Handle response (optional but recommended)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}
	fmt.Printf("Response: %s\n", body)

	// Unmarshal the response into the struct
	var checkoutRes MNOCheckoutResponse
	if err := json.Unmarshal(body, &checkoutRes); err != nil {
		log.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Return the parsed auth token
	return checkoutRes.Success, checkoutRes.TransactionId
}

func AzamPayCallbackHandler(c *gin.Context) {
	// Extract the necessary fields from the request body (or query parameters, etc.)
	var body struct {
		AccountNumber     string `json:"msisdn"`
		Amount            string `json:"amount"`
		Message           string `json:"message"`
		UtilityRef        string `json:"utilityRef"`
		Operator          string `json:"operator"`
		Reference         string `json:"reference"`
		TransactionStatus string `json:"transactionstatus"`
		SubmerchantAcc    string `json:"submerchantAcc"`
		FspReferenceId    string `json:"fspReferenceId"`
	}

	// Bind JSON input to the struct
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})
		return
	}

	// Respond with success
	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})

}
