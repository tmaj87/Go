package main

import (
    "crypto/rand"
    "encoding/base64"
    "errors"
    "net/http"
    "strconv"
    
    "github.com/gin-gonic/gin"
)

// generatePassword generates a random password of specified length
func generatePassword(length int) (string, error) {
    if length < 1 || length > 64 {
        return "", errors.New("invalid length")
    }

    const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*()+_-=?:{}|<>~"
    byteLength := length
    
    randomBytes := make([]byte, byteLength)
    _, err := rand.Read(randomBytes)
    if err != nil {
        return "", errors.New("failed to generate random bytes")
    }

    encodedBytes := base64.URLEncoding.EncodeToString(randomBytes)
    
    // Truncate or pad the password to match the required length
    if len(encodedBytes) > byteLength {
        encodedBytes = encodedBytes[:byteLength]
    } else {
        for len(encodedBytes) < byteLength {
            encodedBytes += letters
        }
        encodedBytes = encodedBytes[:byteLength] // Truncate after padding
    }

    return encodedBytes, nil
}

func main() {
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

    router.Run(":8080")
}