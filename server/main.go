package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"flag"
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

type application struct {
	// authentication between client and server
	auth struct {
		username string
		password string
	}

	// authentication between server and artifactory
	artifactory struct {
		username string
		password string
	}
}

var app application

func init() {
	app.auth.username = os.Getenv("AUTH_USERNAME")
	app.auth.password = os.Getenv("AUTH_PASSWORD")
	if app.auth.username == "" {
		log.Fatal("basic auth username is missing")
	}

	if app.auth.password == "" {
		log.Fatal("basic auth password is missing")
	}

	app.artifactory.username = os.Getenv("ARTIFACTORY_USERNAME")
	app.artifactory.password = os.Getenv("ARTIFACTORY_PASSWORD")
	if app.artifactory.username == "" {
		fmt.Println("artifactory username is missing")
	}

	if app.artifactory.password == "" {
		fmt.Println("artifactory password is missing")
	}
}

func main() {
	var pemPath string
	flag.StringVar(&pemPath, "pem", "../certs/localhost.crt", "path to pem file")
	var keyPath string
	flag.StringVar(&keyPath, "key", "../certs/localhost.key", "path to key file")

	mux := http.NewServeMux()
	addHTTPRoutes(mux)

	server := &http.Server{
		Addr:         ":6000",
		Handler:      mux,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	log.Printf("Starting server on %s", server.Addr)

	err := server.ListenAndServeTLS(pemPath, keyPath)
	log.Fatal(err)
}

func addHTTPRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/config", app.configHandler)
	mux.HandleFunc("/test", app.testHandler)
	mux.HandleFunc("/basic_auth", app.basicAuth(app.protectedBasicAuthHandler))
	mux.HandleFunc("/jwt_auth", app.jwtAuth(app.protectedJWTHandler))
}

// unprotected config handler
func (app *application) configHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "This is the unprotected config handler")
}

// unprotected test handler
func (app *application) testHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "This is the unprotected test handler")
}

// Basic Auth Handlers
func (app *application) basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if ok {
			usernameHash := sha256.Sum256([]byte(username))
			passwordHash := sha256.Sum256([]byte(password))
			expectedUsernameHash := sha256.Sum256([]byte(app.auth.username))
			expectedPasswordHash := sha256.Sum256([]byte(app.auth.password))

			usernameMatch := (subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1)
			passwordMatch := (subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1)

			if usernameMatch && passwordMatch {
				next.ServeHTTP(w, r)
				return
			}
		}

		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}

func (app *application) protectedBasicAuthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "This is the protected handler")
	fmt.Fprintf(w, "Basic Auth Passed")

	if app.artifactory.username != "" && app.artifactory.password != "" {
		fetchArtifact(w)
	}
}

// JWT Auth Handlers
func (app *application) jwtAuth(endpoint func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header["Token"] != nil {
			token, err := jwt.Parse(r.Header["Token"][0], func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("Token Error")
				}
				return mySigningKey, nil
			})

			if err != nil {
				fmt.Fprintf(w, err.Error())
			}

			if token.Valid {
				endpoint(w, r)
			}
		} else {
			fmt.Fprintf(w, "Not Authorized")
		}
	})
}

func (app *application) protectedJWTHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "This is the unprotected JWT handler")
	fmt.Fprintf(w, "Super Secret Information")
	if app.artifactory.username != "" && app.artifactory.password != "" {
		fetchArtifact(w)
	}
}

// Helper function
func fetchArtifact(w http.ResponseWriter) {
	// fetch artifact from repo
	username := app.artifactory.username
	password := app.artifactory.password
	server_url := "software.128technology.com"
	file_name := "artifactory/rpm-128t-alpha-local/repodata/repomd.xml"

	client := http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest(http.MethodGet, "https://"+server_url+"/"+file_name, http.NoBody)
	if err != nil {
		log.Fatal(err)
	}
	req.SetBasicAuth(username, password)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}

	defer res.Body.Close()

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintln(w, string(resBody))
}
