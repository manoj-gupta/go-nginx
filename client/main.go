package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

var mySigningKey = []byte("QuickBrownFoxJumpsOverTheLazyDog") // testing
// var mySigningKey = []byte(os.Getenv("MY_JWT_SIGNING_KEY")) // production

const port = "6000"

func main() {
	client := http.Client{Timeout: 5 * time.Second}

	// Open handler (without authentication)
	getOpen(client, buildURL("config"))
	getOpen(client, buildURL("test"))

	// Basic Authentication
	getProtectedBasicAuth(client, buildURL("basic_auth"))

	// JWT Authentication
	getProtectedJWT(client, buildURL("jwt_auth"))
}

func getOpen(client http.Client, url string) {
	req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	if err != nil {
		log.Fatal(err)
	}

	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Status: %d\n", res.StatusCode)
	fmt.Printf("Body: %s\n", string(resBody))
}

func getProtectedBasicAuth(client http.Client, url string) {
	req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	if err != nil {
		log.Fatal(err)
	}

	username := os.Getenv("AUTH_USERNAME")
	password := os.Getenv("AUTH_PASSWORD")
	req.SetBasicAuth(username, password)

	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Status: %d\n", res.StatusCode)
	fmt.Printf("Body: %s\n", string(resBody))
}

func getProtectedJWT(client http.Client, url string) {
	validToken, err := GenerateJWT()
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
		return
	}

	req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Generated Token:%s\n", validToken)
	req.Header.Set("Token", validToken)
	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
		return
	}

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
		return
	}

	fmt.Printf("Status: %d\n", res.StatusCode)
	fmt.Printf("Body: %s\n", string(resBody))
}

func GenerateJWT() (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["authorized"] = true
	claims["user"] = "Manoj Gupta"
	claims["exp"] = time.Now().Add(time.Minute * 30).Unix()

	tokenString, err := token.SignedString(mySigningKey)
	if err != nil {
		fmt.Printf("Error generating token string: %s", err.Error())
		return "", err
	}

	return tokenString, nil
}

func buildURL(handler string) string {
	return "https://localhost:" + port + "/" + handler
}
