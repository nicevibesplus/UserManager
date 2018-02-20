package main

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type config struct {
	LDAPAdmin       string
	LDAPPass        string
	PrivateKey      string
	PublicKey       string
	LDAPServer      string
	LDAPPort        string
	LDAPBaseDN      string
	LDAPAdminfilter string
}

type UserCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Fs string `json:"fs"`
}

var (
	verifyKey     *rsa.PublicKey
	signKey       *rsa.PrivateKey
	configuration config
)

func main() {
	ReadConfig()
	initKeys()

	router := httprouter.New()
	router.HandlerFunc("POST", "/login", Login)
	router.Handler("POST", "/users/add", ValidateTokenMiddleware(UsersAdd()))
	router.Handler("POST", "/users/remove", ValidateTokenMiddleware(UsersRemove()))
	router.Handler("POST", "/users/removeFromList", ValidateTokenMiddleware(RemoveUserFromGroup()))
	router.Handler("GET", "/users/list", ValidateTokenMiddleware(UsersList()))
	router.Handler("POST", "/groups/add", ValidateTokenMiddleware(GroupsAdd()))
	router.Handler("POST", "/groups/remove", ValidateTokenMiddleware(GroupsRemove()))
	router.Handler("GET", "/groups/list", ValidateTokenMiddleware(GroupsList()))

	// Set CORS Headers
	handler := cors.Default().Handler(router)
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST"},
	})

	// Start Server.
	log.Fatal(http.ListenAndServe(":8081", c.Handler(handler)))
}

func Login(w http.ResponseWriter, r *http.Request) {
	var user UserCredentials
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusForbidden)
		fmt.Printf("Error in request", w)
		return
	}

	// LDAP Authentication
	authenticated := LDAPAuthenticateAdmin(user)
	if authenticated == false {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Invalid Credentials"))
		return
	}

	token := jwt.New(jwt.SigningMethodRS256)
	claims := make(jwt.MapClaims)

	claims["exp"] = time.Now().Add(time.Minute * time.Duration(10)).Unix()
	claims["iat"] = time.Now().Unix()
	token.Claims = claims

	tokenString, err := token.SignedString(signKey)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Error while signing the token")
		Fail(err)
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(tokenString))
}

func ReadConfig() {
	file, err := ioutil.ReadFile("config.json")
	Fail(err)

	err = json.Unmarshal(file, &configuration)
	Fail(err)
}

// Code from http://www.giantflyingsaucer.com/blog/?p=5994
func initKeys() {
	signBytes, err := ioutil.ReadFile(configuration.PrivateKey)
	Fail(err)

	signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	Fail(err)

	verifyBytes, err := ioutil.ReadFile(configuration.PublicKey)
	Fail(err)

	verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
}

// Code from http://www.giantflyingsaucer.com/blog/?p=5994
func ValidateTokenMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := request.ParseFromRequest(r, request.AuthorizationHeaderExtractor,
			func(token *jwt.Token) (interface{}, error) {
				return verifyKey, nil
			})

		if err == nil {
			if token.Valid {
				handler.ServeHTTP(w, r)
			} else {
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprint(w, "Token is not valid")
				return
			}
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, "Unauthorized access to this resource")
			return
		}
	})
}
