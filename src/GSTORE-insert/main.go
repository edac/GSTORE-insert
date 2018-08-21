package main

import (
	"fmt"
	// _ "github.com/lib/pq"
	// "database/sql"
	"io/ioutil"
	"log"
	"os"
	"time"
)

//VERSION is and exported variable so the handelers can use it.
var VERSION string

//CODE ... I don't remember. Do we even need this?
var CODE string

// CODENAME is like a major version string
var CODENAME string

//altpath is used if you need an alternate path for some web servers.
var altpath string

//dbuser pass and name and secret to be pulled from config
var dbuser string
var dbpass string
var dbname string

func main() {
	VERSION = "0.0"
	CODENAME = "sledge"
	argsWithoutProg := os.Args[1:]
	fmt.Println(argsWithoutProg)
	//If args given...
	if len(argsWithoutProg) > 0 {
		if argsWithoutProg[0] == "install" {
			if _, err := os.Stat("/etc/GSTORE-insert/"); os.IsNotExist(err) {
				pathErr := os.MkdirAll("/etc/GSTORE-insert/", 0777)
				if pathErr != nil {
					fmt.Println(pathErr)
				}
				d1 := []byte("#Log files location\nLogDir = \"/var/log/\"\n\n#the server socket info\nIP = \"localhost\"\nPort = \"888\"")
				err := ioutil.WriteFile("/etc/GSTORE-insert/GSTORE-insert.conf", d1, 0644)
				if err != nil {
					fmt.Println(err)
				}
				os.OpenFile("/var/log/GSTORE-insert.log", os.O_RDONLY|os.O_CREATE, 0666)
			}
		} else {
			fmt.Println("Unknown param")
		}
	} else { //If no args given...

		var configf = ReadConfig() //this is in config.go

		dbuser = configf.DBUser
		dbpass = configf.DBPass
		dbname = configf.DBName
		fmt.Println("b")
		//Bulk Insert run every minute..
		ticker := time.NewTicker(time.Minute * 1)
		go func() {
			for t := range ticker.C {
				BulkInsert()
				_ = t
			}
		}()

		go forever()
		select {} // block forever

	}
}

//Append is a function for appending slices
func Append(slice []string, items ...string) []string {
	for _, item := range items {
		slice = Extend(slice, item)
	}
	return slice
}
func forever() {
	for {
		// fmt.Printf("%v+\n", time.Now())
		// time.Sleep(time.Second)
		time.Sleep(time.Second * 30)
	}
}

//Extend is an easy wat to grow a slice.
func Extend(slice []string, element string) []string {
	n := len(slice)
	if n == cap(slice) {
		// Slice is full; must grow.
		// We double its size and add 1, so if the size is zero we still grow.
		newSlice := make([]string, len(slice), 2*len(slice)+1)
		copy(newSlice, slice)
		slice = newSlice
	}
	slice = slice[0 : n+1]
	slice[n] = element
	return slice
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func logErr(err error) {
	if err != nil {
		log.Println(err)
	}
}

//BulkInsert check db for DIPS, and if it finds one that has the corect status it runs it.
func BulkInsert() {
	fmt.Printf("%v+\n", time.Now())

	fmt.Println("BulkInsert Running")
}
