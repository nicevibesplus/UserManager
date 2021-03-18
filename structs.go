package main

// ServerConfig holds all configuration options parsed from config.conf file
type ServerConfig struct {
	ServerBindAddr   string
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
}

// User is the internal Representation of User to be added/removed/edited
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Fs       string `json:"fs"`
	Group    string `json:"groupname"`
}

// Group is the internal Representation of Group to be added/removed
type Group struct {
	Name string `json:"groupname"`
}
