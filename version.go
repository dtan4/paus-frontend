package main

import (
	"fmt"
)

var (
	BuildTime string
	Revision  string
	Version   string
)

func printVersion() {
	fmt.Println("Version:   " + Version)
	fmt.Println("Revision:  " + Revision)
	fmt.Println("BuildTime: " + BuildTime)
}
