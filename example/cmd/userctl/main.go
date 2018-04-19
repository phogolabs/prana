package main

import (
	"fmt"
	"time"

	randomdata "github.com/Pallinder/go-randomdata"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/phogolabs/oak/example"
	validator "gopkg.in/go-playground/validator.v9"

	"github.com/apex/log"
	"github.com/phogolabs/oak"
	"github.com/phogolabs/oak/example/database/model"
	"github.com/phogolabs/parcel"
	lk "github.com/ulule/loukoum"
)

func main() {
	gateway, err := oak.OpenURL("sqlite3://oak.db")
	if err != nil {
		log.WithError(err).Fatal("Failed to open database connection")
	}
	defer gateway.Close()

	if err = oak.LoadSQLCommandsFrom(parcel.Root("database/script")); err != nil {
		log.WithError(err).Fatal("Failed to load script")
	}

	if err = oak.Migrate(gateway, parcel.Root("database/migration")); err != nil {
		log.WithError(err).Fatal("Failed to load script")
	}

	if err = create(gateway); err != nil {
		log.WithError(err).Fatal("Failed to generate users")
	}

	users := []model.User{}

	if err = gateway.Select(&users, oak.Command("show-users")); err != nil {
		log.WithError(err).Fatal("Failed to select all users")
	}

	show(users)
}

func create(gateway *oak.Gateway) error {
	for i := 0; i < 10; i++ {
		var lastName interface{}

		if i%2 == 0 {
			lastName = randomdata.LastName()
		}

		query := lk.Insert("users").
			Set(
				lk.Pair("id", time.Now().UnixNano()),
				lk.Pair("first_name", randomdata.FirstName(randomdata.Male)),
				lk.Pair("last_name", lastName),
			)

		if _, err := gateway.Exec(query); err != nil {
			return err
		}
	}

	return nil
}

func show(users []model.User) {
	validate := validator.New()

	for _, user := range users {
		if err := validate.Struct(user); err != nil {
			log.WithError(err).Error("Failed to validate user")
			continue
		}

		fmt.Printf("User ID: %v\n", user.ID)
		fmt.Printf("First Name: %v\n", user.FirstName)

		if user.LastName.Valid {
			fmt.Printf("Last Name: %v\n", user.LastName.String)
		} else {
			fmt.Println("Last Name: null")
		}

		fmt.Println("---")
	}
}
