package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"shive/models"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

/**
IF: TEST ARE FAILING IN GITHUB ACTIONS, UPDATE IP ADDRESS ON THE `Network Access` TAB.
LOOK AT THIS LINK:
	- https://stackoverflow.com/questions/75839566/why-cant-i-connect-to-mongodb-atlas-with-github-actions
 	- https://medium.com/@clemensstich/how-to-use-mongodb-in-github-actions-bf24a0d9adf3
*/

type LoginResponse struct {
	ID           string    `json:"ID"`
	Name         string    `json:"name"`
	Username     string    `json:"username"`
	Password     *string   `json:"password"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	UserType     string    `json:"user_type"`
	RefreshToken *string   `json:"refresh_token"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	UserID       string    `json:"user_id"`
	Status       int       `json:"status"`
	Message      string    `json:"message"`
}

type TestUser struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`

	UserType string `json:"user_type"`
}

var baseURL = os.Getenv("API_URL")

func TestAPIEndpoints(t *testing.T) {
	if baseURL == "" {
		baseURL = "http://localhost:9000"
	}
	testUser := TestUser{
		Name:     "Test User",
		Username: "testuser",
		Password: "testpass123",
		Email:    "test@example.com",
		UserType: "ADMIN",
	}

	// {
	// 	"email": "john.doe@example.com",
	// 	"password": "password",
	// 	"name": "John Doe",
	// 	"username": "johndoe",
	// 	"user_type": "USER"
	// }
	// Setup: Start the server and wait for it to be ready
	time.Sleep(2 * time.Second)

	t.Run("Login Flow", func(t *testing.T) {
		// Test Login
		token, userID := testLogin(t, baseURL, testUser)
		assert.NotEmpty(t, token, "Token should not be empty")
		assert.NotEmpty(t, userID, "UserID should not be empty")

		// Test Get User Details
		t.Run("Get User Details", func(t *testing.T) {
			testGetUserDetails(t, baseURL, token, userID, testUser)
		})

	})

}

func testLogin(t *testing.T, baseURL string, user TestUser) (string, string) {
	loginURL := fmt.Sprintf("%s/users/login", baseURL)
	jsonData, _ := json.Marshal(user)

	// Create a new request
	req, err := http.NewRequest("POST", loginURL, bytes.NewBuffer(jsonData))
	assert.NoError(t, err, "Login request should not error")

	// Set the content type header
	req.Header.Set("Content-Type", "application/json")

	// Create a new client
	client := &http.Client{}

	// Send the request
	resp, err := client.Do(req)
	assert.NoError(t, err, "Login request should not error")

	defer resp.Body.Close()

	var loginResp LoginResponse
	err = json.NewDecoder(resp.Body).Decode(&loginResp)
	assert.NoError(t, err, "Should decode response")

	// LOG response if status is not 200
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Login Response:", loginResp)
	}
	// Assert response structure
	assert.NotEmpty(t, loginResp.ID, "ID should not be empty")
	assert.Equal(t, user.Name, loginResp.Name, "Name should match")
	assert.Equal(t, user.Username, loginResp.Username, "Username should match")
	assert.Equal(t, user.Email, loginResp.Email, "Email should match")
	assert.Equal(t, user.UserType, loginResp.UserType, "UserType should match")
	assert.NotEmpty(t, loginResp.Token, "Token should not be empty")
	assert.Nil(t, loginResp.Password, "Password should be null")
	assert.NotEmpty(t, loginResp.UserID, "UserID should not be empty")

	// Assert JWT token format (basic check)
	assert.True(t, strings.HasPrefix(loginResp.Token, "eyJ"), "Token should be in JWT format")

	// Assert timestamps
	assert.False(t, loginResp.UpdatedAt.IsZero(), "UpdatedAt should not be zero")

	// Assert status code
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Should return 200 OK")

	return loginResp.Token, loginResp.UserID
}

func testGetUserDetails(t *testing.T, baseURL string, token string, userID string, testUser TestUser) {
	// Skip this test if token or userID is empty
	if token == "" || userID == "" {
		t.Skip("Skipping user details test due to missing token or userID")
	}

	userDetailsURL := fmt.Sprintf("%s/users/%s", baseURL, userID)
	req, err := http.NewRequest("GET", userDetailsURL, nil)
	assert.NoError(t, err, "User details request should not error")

	req.Header.Set("token", token)
	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err, "User details request should not error")
	defer resp.Body.Close()

	var userDetails models.User
	err = json.NewDecoder(resp.Body).Decode(&userDetails)
	if err != nil {
		t.Fatalf("Failed to decode user details response: %v", err)
	}

	// Null check before dereferencing pointers
	if userDetails.Email == nil || userDetails.Username == nil {
		t.Fatal("Email or Username is nil in response")
	}
	// Assert that the response was successful
	assert.NoError(t, err, "Should decode response")
	// Assert status code
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Should return 200 OK")

	// Assert email and username
	assert.Equal(t, testUser.Email, *userDetails.Email, "Email should match")
	assert.Equal(t, testUser.Username, *userDetails.Username, "Username should match")

}
