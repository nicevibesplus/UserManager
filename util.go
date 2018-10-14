package main

import (
	"log"
	"net/http"
	"errors"
)

func Fail(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}

func parseUser(r *http.Request) (error, UserCredentials){
	r.ParseForm()

	username := r.PostForm.Get("username")
	password := r.PostForm.Get("password")
	fs := r.PostForm.Get("fs")

	if username == "" || password == "" {
		return errors.New("Error parsing User to struct"), UserCredentials{}
	}

	return nil, UserCredentials{username,password, fs}
}

func parseGroupUser(r *http.Request) (error, GroupUser) {
	r.ParseForm()

	user := r.PostForm.Get("username")
	group := r.PostForm.Get("group")

	if user == "" || group == "" {
		return errors.New("Error parsing User to struct"), GroupUser{}
	}

	return nil, GroupUser{user, group}
}