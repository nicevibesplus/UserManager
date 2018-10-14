package main

import (
	"net/http"
	"fmt"
	"strings"
)

type GroupUser struct {
	user string `json:"user"`
	group string `json:"group"`
}

// UsersList returns a List of all LDAP Users
func UsersList() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		users, err := LDAPViewUsers()
		Fail(err)
		userstring := "[" + strings.Join(users, ",") + "]"
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(userstring))
	})
}

// UsersAdd Adds the new User to the Database
func UsersAdd() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err, user := parseUser(r)
		if err != nil {
			w.WriteHeader(400)
			w.Write([]byte("Error parsing Request Body: " + err.Error()))
			return
		}

		// Check if already Registered
		existing, err := pLDAPSearch(
			[]string{"dn"},
			fmt.Sprintf("(&(objectClass=organizationalPerson)(cn=%s))",user.Username),
		)
		if len(existing) != 0 {
			// User already exists in LDAP
			w.WriteHeader(409)
			w.Write([]byte("User with given Username already exists in LDAP"))
			return
		}
		// Add user to LDAP
		err = LDAPAddUser("cn=" + user.Username + ",o=" + user.Fs + "," + configuration.LDAPBaseDN, user)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Error adding user: " + err.Error()))
			return
		}
		w.WriteHeader(200)
	})
}

// UsersRemove Removes the user with specified dn from the Database
func UsersRemove() http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err, user := parseUser(r)
		if err != nil {
			w.WriteHeader(400)
			w.Write([]byte("Error parsing Request Body: " + err.Error()))
			return
		}

		err = LDAPDeleteUser("cn=" + user.Username + ",o=" + user.Fs + "," + configuration.LDAPBaseDN)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Error deleting user: " + err.Error()))
			return
		}
		w.WriteHeader(200)
	})
}

// RemoveUserFromGroup Removes a user from group
func RemoveUserFromGroup() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err, user := parseGroupUser(r)
		if err != nil {
			w.WriteHeader(400)
			w.Write([]byte("Error parsing Request Body: " + err.Error()))
			return
		}
		err = LDAPRemoveUserFromGroup(user.user, user.group, nil)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Error Removing User from Group: " + err.Error()))
			return
		}
		w.WriteHeader(200)
	})
}

func AddUserToGroup() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err, user := parseGroupUser(r)
		if err != nil {
			w.WriteHeader(400)
			w.Write([]byte("Error parsing Request Body: " + err.Error()))
			return
		}
		err = LDAPAddUserToGroup(user.user, user.group)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Error Adding User from Group: " + err.Error()))
			return
		}
		w.WriteHeader(200)
	})

}

// GroupsList lists all LDAP users
func GroupsList() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
}

// GroupsAdd adds a new group to the LDAP directory
func GroupsAdd() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

}

// GroupsRemove removes a group from the LDAP directory
// The admin group cannot be removed
func GroupsRemove() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

}