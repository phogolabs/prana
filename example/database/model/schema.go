// Package model contains an object model of database schema 'default'
// Auto-generated at Thu Apr 19 21:36:35 CEST 2018
package model

import null "gopkg.in/volatiletech/null.v6"

// User represents a data base table 'users'
type User struct {
	// ID represents a database column 'id' of type 'INT PRIMARY KEY NOT NULL'
	ID int `db:"id,primary_key,not_null" json:"id" xml:"id" validate:"required"`

	// FirstName represents a database column 'first_name' of type 'TEXT NOT NULL'
	FirstName string `db:"first_name,not_null" json:"first_name" xml:"first_name" validate:"required"`

	// LastName represents a database column 'last_name' of type 'TEXT NULL'
	LastName null.String `db:"last_name,null" json:"last_name" xml:"last_name" validate:"-"`
}
