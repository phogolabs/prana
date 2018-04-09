// Package contains an object model of database schema 'default'
// Auto-generated at Mon Apr  9 16:19:24 CEST 2018
package model

import null "gopkg.in/volatiletech/null.v6"

// User represents a data base table 'users'
type User struct {
	// Id represents a database column 'id' of type 'INT NOT NULL'
	Id int `db:"id" json:"id" validate:"required"`

	// FirstName represents a database column 'first_name' of type 'TEXT NOT NULL'
	FirstName string `db:"first_name" json:"first_name" validate:"required"`

	// LastName represents a database column 'last_name' of type 'TEXT NULL'
	LastName null.String `db:"last_name" json:"last_name"`
}
