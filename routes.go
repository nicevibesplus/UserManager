package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
)

var userWithName = map[string]struct{}{"username":{}}
var userWithNameGroup = map[string]struct{}{"username":{}, "group": {}}
var userWithNamePassword = map[string]struct{}{"username":{}, "password": {}}
var userWithNamePasswordFs = map[string]struct{}{"username":{}, "password": {}, "fs": {}}

func Login(w http.ResponseWriter, r *http.Request) {
	err, user := parseUser(r, userWithNamePassword)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
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

// UsersList returns a List of all LDAP Users
func UsersList() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		users, err := LDAPViewUsers()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error occured: " + err.Error()))
			return
		}
		userstring := "[" + strings.Join(users, ",") + "]"
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(userstring))
	})
}

// UsersAdd Adds the new User to the Database
func UsersAdd() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err, user := parseUser(r, userWithNamePasswordFs)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Error parsing Request Body: " + err.Error()))
			return
		}

		// Check if already Registered
		existing, err := pLDAPSearch(
			[]string{"dn"},
			fmt.Sprintf("(&(objectClass=organizationalPerson)(cn=%s))", user.Username),
		)
		if len(existing) != 0 {
			// User already exists in LDAP
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte("User with given Username already exists in LDAP"))
			return
		}
		// Add user to LDAP
		err = LDAPAddUser("cn="+user.Username+",o="+user.Fs+","+configuration.LDAPBaseDN, user)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error adding user: " + err.Error()))
			return
		}
		w.WriteHeader(http.StatusOK)
	})
}

// UsersRemove Removes the user with specified dn from the Database
func UsersRemove() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err, user := parseUser(r, userWithName)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Error parsing Request Body: " + err.Error()))
			return
		}

		if user.Username == "admin" {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error deleting user: User is protected by divine spirits."))
			return
		}

		err = LDAPDeleteDN("cn=" + user.Username + ",o=" + user.Fs + "," + configuration.LDAPBaseDN)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error deleting user: " + err.Error()))
			return
		}
		w.WriteHeader(http.StatusOK)
	})
}

// RemoveUserFromGroup Removes a user from group
func RemoveUserFromGroup() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err, user := parseUser(r, userWithNameGroup)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Error parsing Request Body: " + err.Error()))
			return
		}
		err = LDAPRemoveUserFromGroup(user.Username, user.Group)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error Removing User from Group: " + err.Error()))
			return
		}
		w.WriteHeader(http.StatusOK)
	})
}

func AddUserToGroup() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err, user := parseUser(r, userWithNameGroup)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Error parsing Request Body: " + err.Error()))
			return
		}
		err = LDAPAddUserToGroup(user.Username, user.Group)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error Adding User from Group: " + err.Error()))
			return
		}
		w.WriteHeader(http.StatusOK)
	})
}

func UsersChangePassword() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err, user := parseUser(r, userWithNamePassword)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Error parsing Request Body: " + err.Error()))
			return
		}

		// Check if already Registered
		existing, err := pLDAPSearch(
			[]string{"dn"},
			fmt.Sprintf("(&(objectClass=organizationalPerson)(cn=%s))", user.Username),
		)
		if len(existing) != 1 {
			// User doesn't exist in LDAP
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("User with given Username does not exist in LDAP"))
			return
		}

		err = LDAPChangeUserPassword(user.Username, user.Password)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error changing password: " + err.Error()))
			return
		}
		w.WriteHeader(http.StatusOK)
	})
}

// GroupsList lists all LDAP users
func GroupsList() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		groups, err := LDAPViewGroups()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error occured: " + err.Error()))
			return
		}
		groupstring := "[" + strings.Join(groups, ",") + "]"
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(groupstring))
	})
}

// GroupsAdd adds a new group to the LDAP directory
func GroupsAdd() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err, group := parseGroup(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Error parsing Request Body: " + err.Error()))
			return
		}

		// Check if already Registered
		existing, err := pLDAPSearch(
			[]string{"dn"},
			fmt.Sprintf("(&(objectClass=groupOfNames)(cn=%s))", group),
		)
		if len(existing) != 0 {
			// Already exists in LDAP
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte("Group with given name already exists in LDAP"))
			return
		}
		// Add user to LDAP
		err = LDAPAddGroup("cn=" + group + "," + configuration.LDAPBaseDN)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error adding Group: " + err.Error()))
			return
		}
		w.WriteHeader(http.StatusOK)
	})

}

// GroupsRemove removes a group from the LDAP directory
// The admin group cannot be removed
func GroupsRemove() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err, group := parseGroup(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Error parsing Request Body: " + err.Error()))
			return
		}

		if group == "admins" {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error deleting Group: admin group cannot be deleted"))
			return
		}
		err = LDAPDeleteDN("cn=" + group + "," + configuration.LDAPBaseDN)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error deleting Group: " + err.Error()))
			return
		}
		w.WriteHeader(http.StatusOK)
	})
}
