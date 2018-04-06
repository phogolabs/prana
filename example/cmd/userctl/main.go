package main

import (
	"bytes"
	"fmt"

	_ "github.com/mattn/go-sqlite3"

	"github.com/apex/log"
	"github.com/phogolabs/gom"
	"github.com/phogolabs/gom/example"
	"github.com/phogolabs/gom/example/database/model"
	lk "github.com/ulule/loukoum"
)

func main() {
	script, err := example.Asset("database/script/20180406191137.sql")
	if err != nil {
		log.WithError(err).Fatal("Failed to load embedded resource")
	}

	if err := gom.Load(bytes.NewBuffer(script)); err != nil {
		log.WithError(err).Fatal("Failed to load script")
	}

	gateway, err := gom.Open("sqlite3", "./gom.db")

	if err != nil {
		log.WithError(err).Fatal("Failed to open database connection")
	}

	defer gateway.Close()

	query := lk.Insert("users").
		Set(
			lk.Pair("id", 1),
			lk.Pair("first_name", "John"),
			lk.Pair("last_name", "Doe"),
		)

	if _, err = gateway.Exec(query); err != nil {
		log.WithError(err).Fatal("Failed to insert new user")
	}

	users := []model.User{}

	if err = gateway.Select(&users, gom.Command("show-users")); err != nil {
		log.WithError(err).Fatal("Failed to select all users")
	}

	for _, user := range users {
		fmt.Println(user.Id, user.FirstName, user.LastName)
	}
}
