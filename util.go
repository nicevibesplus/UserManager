package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

func parseUser(r *http.Request, required map[string]struct{}) (error, User) {
	var uc User
	if strings.Contains(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
		r.ParseForm()
		username := r.PostForm.Get("username")
		password := r.PostForm.Get("password")
		fs := r.PostForm.Get("fs")
		group := r.PostForm.Get("groupname")
		uc = User{username, password, fs, group}
	} else if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		decoder := json.NewDecoder(r.Body)
		decoder.Decode(&uc)
	} else {
		return errors.New("could not parse user (invalid format)"), User{}
	}

	// Validate user
	if _, ok := required["username"]; ok && uc.Username == "" {
		return errors.New("could not parse user. (no username supplied)"), User{}
	}
	if _, ok := required["password"]; ok && uc.Password == "" {
		return errors.New("could not parse user. (no password supplied)"), User{}
	}
	if _, ok := required["fs"]; ok && uc.Fs == "" {
		return errors.New("could not parse user. (no fs supplied)"), User{}
	}
	if _, ok := required["group"]; ok && uc.Group== "" {
		return errors.New("could not parse user. (no group supplied)"), User{}
	}

	return nil, uc;
}

func parseGroup(r *http.Request) (error, string) {
	var name string
	if strings.Contains(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
		r.ParseForm()
		name = r.PostForm.Get("groupname")
	} else if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		var uc Group
		decoder := json.NewDecoder(r.Body)
		decoder.Decode(&uc)
		name = uc.Name
	} else {
		return errors.New("could not parse group (invalid format)"), ""
	}

	if name == "" {
		return errors.New("could not parse user (no groupname supplied)"), ""
	}
	return nil, name
}
