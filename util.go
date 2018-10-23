package main

import (
	"log"
	"net/http"
	"errors"
	"strings"
	"encoding/json"
)

func Fail(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}

func parseUser(r *http.Request) (error, UserCredentials){

	if strings.Contains(r.Header.Get("Content-Type"),"application/x-www-form-urlencoded") {
		r.ParseForm()
		username := r.PostForm.Get("username")
		password := r.PostForm.Get("password")
		fs := r.PostForm.Get("fs")
		if username == "" || password == "" {
			return errors.New("Error parsing User to struct"), UserCredentials{}
		}
		return nil, UserCredentials{username,password, fs}
	} else if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		var uc UserCredentials
		decoder := json.NewDecoder(r.Body)
		decoder.Decode(&uc)
		if uc.Username == "" || uc.Password == "" {
			return errors.New("Error parsing User to struct"), UserCredentials{}
		}
		return nil, uc
	} else {
		return errors.New("Error parsing User to struct"), UserCredentials{}
	}
}

func parseGroupUser(r *http.Request) (error, GroupUser) {

	if strings.Contains(r.Header.Get("Content-Type"),"application/x-www-form-urlencoded") {
		r.ParseForm()
		user := r.PostForm.Get("username")
		group := r.PostForm.Get("group")
		if user == "" || group == "" {
			return errors.New("Error parsing User to struct"), GroupUser{}
		}
		return nil, GroupUser{user,group}
	} else if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		var gc GroupUser
		decoder := json.NewDecoder(r.Body)
		decoder.Decode(&gc)
		if gc.user == "" || gc.group == "" {
			return errors.New("Error parsing User to struct"), GroupUser{}
		}
		return nil, gc
	} else {
		return errors.New("Error parsing User to struct"), GroupUser{}
	}
}