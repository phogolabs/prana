package gom_test

import (
	"fmt"

	"github.com/phogolabs/gom"
	lk "github.com/ulule/loukoum"
)

type User struct {
	ID        int64  `db:"id"`
	FirstName string `db:"last_name"`
	LastName  string `db:"first_name"`
}

func ExampleGatewaySelectOne() {
	gateway, err := gom.Open("sqlite3", "example.db")
	if err != nil {
		fmt.Println(err)
		return
	}

	defer func() {
		if dbErr := gateway.Close(); dbErr != nil {
			fmt.Println(dbErr)
		}
	}()

	query := lk.Select("id", "first_name", "last_name").
		From("users").
		Where(lk.Condition("first_name").Equal("John"))

	user := User{}
	if err := gateway.SelectOne(&user, query); err != nil {
		fmt.Println(err)
	}
}

func ExampleGatewaySelect() {
	gateway, err := gom.Open("sqlite3", "example.db")
	if err != nil {
		fmt.Println(err)
		return
	}

	defer func() {
		if dbErr := gateway.Close(); dbErr != nil {
			fmt.Println(dbErr)
		}
	}()

	query := lk.Select("id", "first_name", "last_name").From("users")
	users := []User{}

	if err := gateway.Select(&users, query); err != nil {
		fmt.Println(err)
	}
}

func ExampleGatewayQueryRow() {
	gateway, err := gom.Open("sqlite3", "example.db")
	if err != nil {
		fmt.Println(err)
		return
	}

	defer func() {
		if dbErr := gateway.Close(); dbErr != nil {
			fmt.Println(dbErr)
		}
	}()

	query := lk.Select("id", "first_name", "last_name").
		From("users").
		Where(lk.Condition("first_name").Equal("John"))

	var row *gom.Row

	row, err = gateway.QueryRow(query)
	if err != nil {
		fmt.Println(err)
		return
	}

	user := User{}
	if err := row.StructScan(&user); err != nil {
		fmt.Println(err)
	}
}

func ExampleGatewayQuery() {
	gateway, err := gom.Open("sqlite3", "example.db")
	if err != nil {
		fmt.Println(err)
		return
	}

	defer func() {
		if dbErr := gateway.Close(); dbErr != nil {
			fmt.Println(dbErr)
		}
	}()

	query := lk.Select("id", "first_name", "last_name").From("users")
	rows, err := gateway.Query(query)

	if err != nil {
		fmt.Println(err)
		return
	}

	user := User{}

	for rows.Next() {
		if err = rows.StructScan(&user); err != nil {
			fmt.Println(err)
			return
		}

		if user.FirstName == "John" {
			fmt.Println(user.LastName)
		}
	}
}

func ExampleGatewayExec() {
	gateway, err := gom.Open("sqlite3", "example.db")
	if err != nil {
		fmt.Println(err)
		return
	}

	defer func() {
		if dbErr := gateway.Close(); dbErr != nil {
			fmt.Println(dbErr)
		}
	}()

	query := lk.Insert("users").
		Set(
			lk.Pair("first_name", "John"),
			lk.Pair("last_name", "Doe"),
		).
		Returning("id")

	if _, err := gateway.Exec(query); err != nil {
		fmt.Println(err)
	}
}

func ExampleGatewayCommand() {
	err := gom.LoadDir("./database/command")

	if err != nil {
		fmt.Println(err)
		return
	}

	cmd := gom.Command("show-sqlite-master")

	gateway, err := gom.Open("sqlite3", "example.db")
	if err != nil {
		fmt.Println(err)
		return
	}

	defer func() {
		if dbErr := gateway.Close(); dbErr != nil {
			fmt.Println(dbErr)
		}
	}()

	if _, err := gateway.Exec(cmd); err != nil {
		fmt.Println(err)
	}
}
