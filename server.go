package main

import (
	"crypto/rsa"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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

func readConfig(conf *ServerConfig) {
	// default values
	conf.ServerBindAddr = ":8443"
	conf.JWTPublicRSAKey = "./keys/jwt.pub"
	conf.JWTPrivateRSAKey = "./keys/jwt.key"
	conf.LDAPServer = "localhost"
	conf.LDAPPort = "389"
	conf.LDAPUserfilter = "(&(objectClass=organizationalPerson)(cn=%s))"

	// load from json
	if file, err := ioutil.ReadFile("config.conf"); err != nil {
		log.Print(err)
		log.Println("couldn't read config file, falling back to defaults + environment variables")
	} else if err = json.Unmarshal(file, &conf); err != nil {
		log.Print(err)
		log.Println("couldn't read config file, falling back to defaults + environment variables")
	}

	if os.Getenv("UM_SERVER_BIND_ADDR") != "" {
		conf.ServerBindAddr = os.Getenv("UM_SERVER_BIND_ADDR")
	}
	if os.Getenv("UM_JWT_PUB") != "" {
		conf.JWTPublicRSAKey = os.Getenv("UM_JWT_PUB")
	}
	if os.Getenv("UM_JWT_PRIV") != "" {
		conf.JWTPrivateRSAKey = os.Getenv("UM_JWT_PRIV")
	}
	if os.Getenv("UM_TLS_CERT") != "" {
		conf.SSLCertificate = os.Getenv("UM_TLS_CERT")
	}
	if os.Getenv("UM_TLS_KEY") != "" {
		conf.SSLKeyFile = os.Getenv("UM_TLS_KEY")
	}
	if os.Getenv("UM_LDAP_ADMIN") != "" {
		conf.LDAPAdmin = os.Getenv("UM_LDAP_ADMIN")
	}
	if os.Getenv("UM_LDAP_PASS") != "" {
		conf.LDAPPass = os.Getenv("UM_LDAP_PASS")
	}
	if os.Getenv("UM_LDAP_BASE_DN") != "" {
		conf.LDAPBaseDN = os.Getenv("UM_LDAP_BASE_DN")
	}
	if os.Getenv("UM_LDAP_SERVER") != "" {
		conf.LDAPServer = os.Getenv("UM_LDAP_SERVER")
	}
	if os.Getenv("UM_LDAP_PORT") != "" {
		conf.LDAPPort = os.Getenv("UM_LDAP_PORT")
	}
	if os.Getenv("UM_LDAP_ADMINFILTER") != "" {
		conf.LDAPAdminfilter = os.Getenv("UM_LDAP_ADMINFILTER")
	}
	if os.Getenv("UM_LDAP_USERFILTER") != "" {
		conf.LDAPUserfilter = os.Getenv("UM_LDAP_USERFILTER")
	}

	// validate required values are set
	if conf.LDAPAdmin == "" {
		log.Fatal("missing required config LDAPAdmin")
	}
	if conf.LDAPPass == "" {
		log.Fatal("missing required config LDAPPass")
	}
	if conf.LDAPBaseDN == "" {
		log.Fatal("missing required config LDAPBaseDN")
	}
	if conf.LDAPAdminfilter == "" {
		log.Fatal("missing required config LDAPAdminfilter")
	}
	// the values below have default values, but we check just in case
	if conf.LDAPServer == "" {
		log.Fatal("missing required config LDAPServer")
	}
	if conf.LDAPPort == "" {
		log.Fatal("missing required config LDAPPort")
	}
	if conf.LDAPUserfilter == "" {
		log.Fatal("missing required config LDAPUserfilter")
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
