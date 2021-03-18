module usermanager

go 1.15

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/didip/tollbooth v4.0.2+incompatible
	github.com/julienschmidt/httprouter v1.3.0
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pkg/errors v0.9.1
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba // indirect
	gopkg.in/asn1-ber.v1 v1.0.0-00010101000000-000000000000 // indirect
	gopkg.in/ldap.v2 v2.5.1
)

// https://github.com/go-asn1-ber/asn1-ber/issues/23
replace gopkg.in/asn1-ber.v1 => github.com/go-asn1-ber/asn1-ber v1.5.3
