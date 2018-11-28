package main

type ServerConfig struct {
	ServerPort       string
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

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Fs       string `json:"fs"`
}

type GroupUser struct {
	User  string `json:"username"`
	Group string `json:"group"`
}
