## Generate passwords with:
    - Uppercase letters (A-Z)
    - Lowercase letters (a-z)
    - Digits (0-9)
    - Special characters (!@#$%^&*()+_-=?:{}|<>~)

Default password length is 32 characters

Maximum allowed password length is 64 characters

## To use this application:

1. Install dependencies:
```bash
go get -u github.com/gin-gonic/gin
```

2. Run the application:
```bash
go run main.go
```

3. Use these endpoints:
- Default password (32 characters):
```bash
curl "http://localhost:8080/generate-password"
```

- Password with specified length (up to 64):
```bash
curl "http://localhost:8080/generate-password?length=50"
```

The application will return a JSON response containing the generated password.