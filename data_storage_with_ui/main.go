package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var (
	dataStore = make(map[string]interface{})
	mu        sync.RWMutex
)

func main() {
	router := gin.Default()

	// enable CORS
	router.Use(cors.Default())

	// store JSON under the provided key
	router.POST("/data", func(c *gin.Context) {
		// get key
		key := c.Query("key")
		if key == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'key' query parameter"})
			return
		}

		// read
		var reqBody map[string]interface{}
		if err := c.ShouldBindJSON(&reqBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// check existence
		value, exists := reqBody["json"]
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'json' property in request body"})
			return
		}

		// validate format
		if _, err := json.Marshal(value); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format in 'json' property"})
			return
		}

		mu.Lock()
		dataStore[key] = value
		mu.Unlock()

		c.JSON(http.StatusOK, gin.H{"message": "Data stored successfully"})
	})

	// return the stored JSON for the given key
	router.GET("/data", func(c *gin.Context) {
		// get key
		key := c.Query("key")
		if key == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'key' query parameter"})
			return
		}

		// retrieve
		mu.RLock()
		value, exists := dataStore[key]
		mu.RUnlock()

		// validate
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Data not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"json": value})
	})

	// returns all existing keys
	router.GET("/list", func(c *gin.Context) {
		mu.RLock()
		keys := make([]string, 0, len(dataStore))
		for k := range dataStore {
			keys = append(keys, k)
		}
		mu.RUnlock()

		c.JSON(http.StatusOK, gin.H{"keys": keys})
	})

	// start server
	err := router.Run(":8080")
	if err != nil {
		fmt.Println("Terminated with error: ", err)
	}
}
