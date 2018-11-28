package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
)

func Fail(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}

func parseUser(r *http.Request) (error, User) {
	if strings.Contains(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
		r.ParseForm()
		username := r.PostForm.Get("username")
		password := r.PostForm.Get("password")
		fs := r.PostForm.Get("fs")
		if username == "" {
			return errors.New("Error parsing User to struct"), User{}
		}
		return nil, User{username, password, fs}
	} else if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		var uc User
		decoder := json.NewDecoder(r.Body)
		decoder.Decode(&uc)
		if uc.Username == "" {
			return errors.New("Error parsing User to struct"), User{}
		}
		return nil, uc
	} else {
		return errors.New("Error parsing User to struct"), User{}
	}
}

func parseGroupUser(r *http.Request) (error, GroupUser) {
	if strings.Contains(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
		r.ParseForm()
		user := r.PostForm.Get("username")
		group := r.PostForm.Get("group")
		if user == "" || group == "" {
			return errors.New("Error parsing User to struct"), GroupUser{}
		}
		return nil, GroupUser{user, group}
	} else if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		var gc GroupUser
		decoder := json.NewDecoder(r.Body)
		decoder.Decode(&gc)
		if gc.User == "" || gc.Group == "" {
			return errors.New("Error parsing User to struct"), GroupUser{}
		}
		return nil, gc
	} else {
		return errors.New("Error parsing User to struct"), GroupUser{}
	}
}

func parseGroup(r *http.Request) (error, string) {
	if strings.Contains(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
		r.ParseForm()
		name := r.PostForm.Get("groupname")
		if name == "" {
			return errors.New("error parsing Group name"), ""
		}
		return nil, name
	} else if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		var uc Group
		decoder := json.NewDecoder(r.Body)
		decoder.Decode(&uc)
		if uc.Name == "" {
			return errors.New("error parsing User to struct"), ""
		}
		return nil, uc.Name
	} else {
		return errors.New("error parsing User to struct"), ""
	}
}
