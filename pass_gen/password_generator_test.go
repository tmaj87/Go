package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
)

// Test the generatePassword function directly.
func TestGeneratePassword(t *testing.T) {
	// Test a valid length.
	length := 16
	pass, err := generatePassword(length)
	if err != nil {
		t.Fatalf("expected no error for valid length, got: %v", err)
	}
	if len(pass) != length {
		t.Errorf("expected password length %d, got %d", length, len(pass))
	}

	// Test length less than 1.
	_, err = generatePassword(0)
	if err == nil {
		t.Errorf("expected error for length < 1, got nil")
	}

	// Test length greater than 64.
	_, err = generatePassword(65)
	if err == nil {
		t.Errorf("expected error for length > 64, got nil")
	}
}

// setupRouter creates a Gin engine with the /generate-password route.
func setupRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/generate-password", func(c *gin.Context) {
		length := c.Query("length")

		var passwordLength int
		if length == "" {
			passwordLength = 32 // default length
		} else {
			pl, err := strconv.Atoi(length)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "invalid length parameter",
				})
				return
			}
			passwordLength = int(pl)
			if passwordLength > 64 {
				passwordLength = 64
			} else if passwordLength < 1 {
				passwordLength = 32 // default if invalid value
			}
		}

		password, err := generatePassword(passwordLength)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to generate password",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"password": password,
		})
	})
	return router
}

// Test the /generate-password endpoint with no query parameter (should use default length 32).
func TestGeneratePasswordEndpoint_DefaultLength(t *testing.T) {
	router := setupRouter()

	req, _ := http.NewRequest("GET", "/generate-password", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, resp.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	password, exists := body["password"]
	if !exists {
		t.Fatal("response does not contain a password field")
	}

	if len(password) != 32 {
		t.Errorf("expected password length 32, got %d", len(password))
	}
}

// Test the endpoint with a valid custom length.
func TestGeneratePasswordEndpoint_CustomLength(t *testing.T) {
	router := setupRouter()
	customLength := 20
	req, _ := http.NewRequest("GET", "/generate-password?length="+strconv.Itoa(customLength), nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, resp.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	password, exists := body["password"]
	if !exists {
		t.Fatal("response does not contain a password field")
	}

	if len(password) != customLength {
		t.Errorf("expected password length %d, got %d", customLength, len(password))
	}
}

// Test the endpoint with an invalid (non-numeric) length parameter.
func TestGeneratePasswordEndpoint_InvalidQuery(t *testing.T) {
	router := setupRouter()
	req, _ := http.NewRequest("GET", "/generate-password?length=invalid", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("expected status code %d for invalid query, got %d", http.StatusBadRequest, resp.Code)
	}
}

// Test the endpoint with a length parameter greater than the allowed maximum.
// According to the implementation, if a user requests a length > 64, it should be set to 64.
func TestGeneratePasswordEndpoint_ExceedingMaxLength(t *testing.T) {
	router := setupRouter()
	req, _ := http.NewRequest("GET", "/generate-password?length=100", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, resp.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	password, exists := body["password"]
	if !exists {
		t.Fatal("response does not contain a password field")
	}

	if len(password) != 64 {
		t.Errorf("expected password length 64 (max), got %d", len(password))
	}
}
