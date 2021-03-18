package main

import (
	"crypto/rsa"
	"crypto/tls"
	"log"
	"net/http"

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

	// Frontend
	router.GET("/", EmbeddedStaticFilesMiddleware)
	router.GET("/static/*filepath", EmbeddedStaticFilesMiddleware)

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

	srv := &http.Server{
		Addr:         configuration.ServerBindAddr,
		Handler:      router,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
	}

	// Start Server.
	if configuration.SSLCertificate == "" && configuration.SSLKeyFile == "" {
		log.Println("listening (http) on", configuration.ServerBindAddr)
		log.Fatal(srv.ListenAndServe())
	} else {
		log.Println("listening (https) on", configuration.ServerBindAddr)
		log.Fatal(srv.ListenAndServeTLS(configuration.SSLCertificate, configuration.SSLKeyFile))
	}
}
