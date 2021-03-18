package main

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/dgrijalva/jwt-go"
)

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

	log.Print(conf)

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
