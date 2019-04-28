package main

import (
	"crypto/rsa"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"

	"github.com/dgrijalva/jwt-go"
	"github.com/didip/tollbooth"
	"github.com/julienschmidt/httprouter"
)

var (
	verifyKey     *rsa.PublicKey
	signKey       *rsa.PrivateKey
	configuration ServerConfig
)

func main() {
	readConfig(&configuration)
	verifyKey, signKey = readJWTKeys(configuration)
	router := httprouter.New()
	ratelimiter := tollbooth.NewLimiter(0.2, nil) // allow one request every 5 seconds per IP

	// API
	router.Handler("POST", "/api/login", tollbooth.LimitFuncHandler(ratelimiter, Login))
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

func readConfig(conf *ServerConfig) {
	file, err := ioutil.ReadFile("config.conf")
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(file, &conf)
	if err != nil {
		log.Fatal(err)
	}
}

func readCert(path string) (result []byte) {
	abspath, err := filepath.Abs(path)
	if err != nil {
		log.Fatal(err)
	}
	result, err = ioutil.ReadFile(abspath)
	if err != nil {
		log.Fatal(err)
	}
	return
}

// Code from http://www.giantflyingsaucer.com/blog/?p=5994
func readJWTKeys(configuration ServerConfig) (verifyKey *rsa.PublicKey, signKey *rsa.PrivateKey) {
	var err error

	if configuration.JWTPrivateRSAKey == "" {
		err = errors.New("missing config key JWTPrivateRSAKey")
		log.Print(err)
	}
	if configuration.JWTPublicRSAKey == "" {
		err = errors.New("missing config key JWTPublicRSAkey")
		log.Print(err)
	}
	if err != nil {
		log.Fatal("incorrect config file")
	}

	signBytes := readCert(configuration.JWTPrivateRSAKey)
	signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		log.Fatal(err)
	}

	verifyBytes := readCert(configuration.JWTPublicRSAKey)
	verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		log.Fatal(err)
	}

	return
}
