package gom_test

import (
	"fmt"

	"github.com/svett/gom"
)

type User struct {
	ID        int64  `db:"id"`
	FirstName string `db:"last_name"`
	LastName  string `db:"first_name"`
}

func ExampleGatewaySelectOne() error {
	gateway, err := gom.Open("sqlite3", "example.db")
	if err != nil {
		return err
	}

	defer func() {
		if dbErr := gateway.Close(); err != nil {
			err = dbErr
		}
	}()

	query := gom.Select("id", "first_name", "last_name").
		From("users").
		Where(gom.Condition("first_name").Equal("John"))

	user := User{}

	err = gateway.SelectOne(&user, query)
	return err
}

func ExampleGatewaySelect() error {
	gateway, err := gom.Open("sqlite3", "example.db")
	if err != nil {
		return err
	}

	defer func() {
		if dbErr := gateway.Close(); err != nil {
			err = dbErr
		}
	}()

	query := gom.Select("id", "first_name", "last_name").From("users")
	users := []User{}

	err = gateway.Select(&users, query)
	return err
}

func ExampleGatewayQueryRow() error {
	gateway, err := gom.Open("sqlite3", "example.db")
	if err != nil {
		return err
	}

	defer func() {
		if dbErr := gateway.Close(); err != nil {
			err = dbErr
		}
	}()

	query := gom.Select("id", "first_name", "last_name").
		From("users").
		Where(gom.Condition("first_name").Equal("John"))

	var row *gom.Row

	row, err = gateway.QueryRow(query)
	if err != nil {
		return err
	}

	user := User{}
	err = row.StructScan(&user)

	return err
}

func ExampleGatewayQuery() error {
	gateway, err := gom.Open("sqlite3", "example.db")
	if err != nil {
		return err
	}

	defer func() {
		if dbErr := gateway.Close(); err != nil {
			err = dbErr
		}
	}()

	query := gom.Select("id", "first_name", "last_name").From("users")
	rows, err := gateway.Query(query)

	if err != nil {
		return err
	}

	user := User{}

	for rows.Next() {
		if err = rows.StructScan(&user); err != nil {
			return err
		}

		if user.FirstName == "John" {
			fmt.Println(user.LastName)
		}
	}

	return err
}

func ExampleGatewayExec() error {
	gateway, err := gom.Open("sqlite3", "example.db")
	if err != nil {
		return err
	}

	defer func() {
		if dbErr := gateway.Close(); err != nil {
			err = dbErr
		}
	}()

	query := gom.Insert("users").
		Set(
			gom.Pair("first_name", "John"),
			gom.Pair("last_name", "Doe"),
		).
		Returning("id")

	_, err = gateway.Exec(query)
	return err
}

func ExampleGatewayCommand() error {
	err := gom.LoadDir("./database/command")

	if err != nil {
		return err
	}

	cmd := gom.Command("show-sqlite-master")

	gateway, err := gom.Open("sqlite3", "example.db")
	if err != nil {
		return err
	}

	defer func() {
		if dbErr := gateway.Close(); err != nil {
			err = dbErr
		}
	}()

	_, err = gateway.Exec(cmd)
	return err
}
