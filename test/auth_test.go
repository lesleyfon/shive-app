package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"shive/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type LoginResponse struct {
	Token   string `json:"token"`
	UserID  string `json:"user_id"`
	Status  int    `json:"status"`
	Message string `json:"message"`
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

	// t.Run("Signup and Login Flow", func(t *testing.T) {
	// 	// Test Signup
	// 	token, userID := testSignup(t, baseURL, testUser)
	// 	assert.NotEmpty(t, token, "Token should not be empty")
	// 	assert.NotEmpty(t, userID, "UserID should not be empty")

	// 	// Test Get User Details
	// 	t.Run("Get User Details", func(t *testing.T) {
	// 		testGetUserDetails(t, baseURL, token, userID)
	// 	})

	// Test Get All Users
	// t.Run("Get All Users", func(t *testing.T) {
	// 	testGetAllUsers(t, baseURL, token)
	// })
	// })
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

// func testSignup(t *testing.T, baseURL string, user TestUser) (string, string) {
// 	signupURL := fmt.Sprintf("%s/users/signup", baseURL)
// 	jsonData, _ := json.Marshal(user)

// 	resp, err := http.Post(signupURL, "application/json", bytes.NewBuffer(jsonData))
// 	assert.NoError(t, err, "Signup request should not error")
// 	defer resp.Body.Close()

// 	var signupResp LoginResponse
// 	err = json.NewDecoder(resp.Body).Decode(&signupResp)
// 	assert.NoError(t, err, "Should decode response")
// 	assert.Equal(t, http.StatusOK, resp.StatusCode, "Should return 200 OK")

// 	return signupResp.Token, signupResp.UserID
// }

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
	// Assert that the response was successful
	assert.NoError(t, err, "Should decode response")
	// Assert status code
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Should return 200 OK")

	return loginResp.Token, loginResp.UserID
}

func testGetUserDetails(t *testing.T, baseURL string, token string, userID string, testUser TestUser) {
	// Format the user details URL
	userDetailsURL := fmt.Sprintf("%s/users/%s", baseURL, userID)

	// Create a new request
	req, err := http.NewRequest("GET", userDetailsURL, nil)
	assert.NoError(t, err, "User details request should not error")

	// Set the token header
	req.Header.Set("token", token)

	// Create a new client
	client := &http.Client{}

	// Send the request
	resp, err := client.Do(req)

	// Assert that the request was successful
	assert.NoError(t, err, "User details request should not error")

	defer resp.Body.Close()

	// Decode the response body into a User struct
	var userDetails models.User
	err = json.NewDecoder(resp.Body).Decode(&userDetails)

	// Assert that the response was successful
	assert.NoError(t, err, "Should decode response")
	// Assert status code
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Should return 200 OK")

	// Assert email and username
	assert.Equal(t, testUser.Email, *userDetails.Email, "Email should match")
	assert.Equal(t, testUser.Username, *userDetails.Username, "Username should match")

}
