package main

import (
	"crypto/rsa"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/julienschmidt/httprouter"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"time"
)

var (
	verifyKey     *rsa.PublicKey
	signKey       *rsa.PrivateKey
	configuration ServerConfig
)

func main() {
	ReadConfig()
	initKeys()
	router := httprouter.New()

	// API
	router.HandlerFunc("POST", "/api/login", Login)
	router.Handler("POST", "/api/users/add", ValidateTokenMiddleware(UsersAdd()))
	router.Handler("POST", "/api/users/remove", ValidateTokenMiddleware(UsersRemove()))
	router.Handler("POST", "/api/users/removeFromGroup", ValidateTokenMiddleware(RemoveUserFromGroup()))
	router.Handler("POST", "/api/users/addToGroup", ValidateTokenMiddleware(AddUserToGroup()))
	router.Handler("POST", "/api/users/changePassword", ValidateTokenMiddleware(UsersChangePassword()))
	router.Handler("GET", "/api/users/list", ValidateTokenMiddleware(UsersList()))

	router.Handler("POST", "/api/groups/add", ValidateTokenMiddleware(GroupsAdd()))
	router.Handler("POST", "/api/groups/remove", ValidateTokenMiddleware(GroupsRemove()))
	router.Handler("GET", "/api/groups/list", ValidateTokenMiddleware(GroupsList()))

	// Frontend
	router.ServeFiles("/static/*filepath", http.Dir("public/static"))
	router.Handler("GET", "/", http.FileServer(http.Dir("public")))

	srv := &http.Server{
		Addr:         ":" + configuration.ServerPort,
		Handler:      router,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
	}

	// Start Server.
	log.Fatal(srv.ListenAndServeTLS(configuration.SSLCertificate, configuration.SSLKeyFile))
}

func Login(w http.ResponseWriter, r *http.Request) {
	err, user := parseUser(r)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		fmt.Printf("Error in request", err)
		return
	}

	// LDAP Authentication
	authenticated, err := LDAPAuthenticateAdmin(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Error while signing the token")
		w.Write([]byte("Error occured: " + err.Error()))
	}
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
		w.Write([]byte("Error occured: " + err.Error()))
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(tokenString))
}

func ReadConfig() {
	file, err := ioutil.ReadFile("config.conf")
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(file, &configuration)
	if err != nil {
		log.Fatal(err)
	}
}

// Code from http://www.giantflyingsaucer.com/blog/?p=5994
func initKeys() {

	signBytes, err := ioutil.ReadFile(configuration.JWTPrivateRSAKey)
	if err != nil {
		log.Fatal(err)
	}

	signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		log.Fatal(err)
	}

	verifyBytes, err := ioutil.ReadFile(configuration.JWTPublicRSAKey)
	if err != nil {
		log.Fatal(err)
	}

	verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		log.Fatal(err)
	}
}

// Code from http://www.giantflyingsaucer.com/blog/?p=5994
func ValidateTokenMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := request.ParseFromRequest(r, request.AuthorizationHeaderExtractor,
			func(token *jwt.Token) (interface{}, error) {
				// Don't forget to validate the alg is what you expect:
				if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
					return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
				}
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
