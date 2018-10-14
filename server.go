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
	"crypto/tls"
)

type config struct {
	ServerPort		 string
	LDAPAdmin        string
	LDAPPass         string
	JWTPrivateRSAKey string
	JWTPublicRSAKey  string
	SSLCertificate   string
	SSLKeyFile       string
	LDAPServer       string
	LDAPPort         string
	LDAPBaseDN       string
	LDAPAdminfilter  string
	LDAPUserfilter   string
	LDAPUserGroups   string
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

	// API
	router.HandlerFunc("POST", "/api/login", Login)
	router.Handler("POST", "/api/users/add", ValidateTokenMiddleware(UsersAdd()))
	router.Handler("POST", "/api/users/remove", ValidateTokenMiddleware(UsersRemove()))
	router.Handler("POST", "/api/users/removeFromGroup", ValidateTokenMiddleware(RemoveUserFromGroup()))
	router.Handler("POST", "/api/users/addToGroup", ValidateTokenMiddleware(AddUserToGroup()))
	router.Handler("GET", "/api/users/list", ValidateTokenMiddleware(UsersList()))
	router.Handler("POST", "/api/groups/add", ValidateTokenMiddleware(GroupsAdd()))
	router.Handler("POST", "/api/groups/remove", ValidateTokenMiddleware(GroupsRemove()))
	router.Handler("GET", "/api/groups/list", ValidateTokenMiddleware(GroupsList()))

	// Frontend
	router.ServeFiles("/static/*filepath", http.Dir("public/static"))
	router.Handler("GET", "/", http.FileServer(http.Dir("public")))

	// Set CORS Headers
	handler := cors.Default().Handler(router)
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST"},
	})

	srv := &http.Server{
		Addr:         ":" + configuration.ServerPort,
		Handler:      c.Handler(handler),
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
	file, err := ioutil.ReadFile("config.conf")
	Fail(err)

	err = json.Unmarshal(file, &configuration)
	Fail(err)
}

// Code from http://www.giantflyingsaucer.com/blog/?p=5994
func initKeys() {

	signBytes, err := ioutil.ReadFile(configuration.JWTPrivateRSAKey)
	Fail(err)

	signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	Fail(err)

	verifyBytes, err := ioutil.ReadFile(configuration.JWTPublicRSAKey)
	Fail(err)

	verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
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
