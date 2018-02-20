package main

import "log"

func Fail(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}
