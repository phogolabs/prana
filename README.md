# OAK

[![Documentation][godoc-img]][godoc-url]
![License][license-img]
[![Build Status][travis-img]][travis-url]
[![Coverage][codecov-img]][codecov-url]
[![Go Report Card][report-img]][report-url]

*Golang Database Manager*

[![OAK][oak-img]][oak-url]

## Overview

OAK is a package for rapid application development with relational databases in
Golang.  It has a command line interface that provides:

- SQL Migrations
- Embedded SQL Scripts
- Model generation from SQL schema

## Installation

```console
$ go get -u github.com/phogolabs/oak
$ go install github.com/phogolabs/oak/cmd/oak
```

## Introduction

### SQL Migrations

The SQL Migration are using SQL command API under the hood. Each migration is a
SQL script that contains `up` and `down` commands.

In order to prepare the project for migration, you have to set it up:

```console
$ oak migration setup
```

Then you can create a migration with the following command:

```console
$ oak migration create schema
```

The command will create the following migration file in `/database/migration`:

```console
$ tree database

database/
└── migration
    ├── 00060524000000_setup.sql
    └── 20180329162010_schema.sql
```

The `20180329162010_schema.sql` migration has similar to example below format:

```sql
-- Auto-generated at Thu Mar 29 16:20:10 CEST 2018
-- Please do not change the name attributes

-- name: up
CREATE TABLE users (
  id INT PRIMARY KEY NOT NULL,
  first_name TEXT NOT NULL,
  last_name TEXT
);

-- name: down
DROP TABLE IF EXISTS users;
```

You can run the migration with the following command:

```console
$ oak migration run
```

If you want to rollback the migration you have to revert it:

```console
$ oak migration revert
```

## SQL Schema and Code Generation

Let's assume that we want to generate a mode for the `users` table.

You can use the `oak` command line interface to generate a package that
contains Golang structs, which represents each table from the desired schema.

For that purpose you should call the following subcommand:

```bash
$ oak model sync
```

By default the command will place the generated code in single `schema.go` file
in `$PWD/database/model` package for the default database schema. Any other
schemas will be placed in the same package but in separate files. You can
control the behavior by passing `--keep-schema` flag which will cause each
schema to be generated in own package under the `/$PWD/database/model` package.

You can print the source code without generating a package by executing the
following command:

```bash
$ oak model print
```

Note that you can specify the desired schema or tables by providing the correct
arguments.

If you pass `--extra-tag` argument, you can specify which tag to be included in
your final result. Supported extra tags are:

- [json](https://golang.org/pkg/encoding/json/) that can be recognized by Golang
- [xml](https://golang.org/pkg/encoding/xml/) tag used by Golang to marshal fields in XML
- [validate](https://github.com/go-playground/validator/blob/v9/_examples/simple/main.go#L11) to validate field by [validator](https://github.com/go-playground/validator) package

The model representation of the users table is:

```golang
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
```

Note that the code generation depends on two packages. In order to produce a
source code that compiles you should have in your `$GOPATH/src` directory
installed:

- [go.uuid](https://github.com/satori/go.uuid) package
- [null](https://github.com/guregu/null) package

The generated `db` tag is recognized by
[parcel.Gateway](https://godoc.org/github.com/phogolabs/oak#Gateway) as well as
[sqlx](https://github.com/jmoiron/sqlx).

If you wan to generate model to work with [gorm](http://gorm.io), you should
pass `--orm-tag gorm`. Note that constraints like unique or indexes are not
included for now.

### SQL Queries with Loukoum

Gateway API facilitates object relation mapping and query building by using
[loukoum](loukoum-url) and [sqlx][sqlx-url]. Before start working you should
import the desired packages:

```golang
import (
  lk "github.com/ulule/loukoum"
  "github.com/phogolabs/oak"
)
```

Let's first establish the connection:

```golang
gateway, err := oak.Open("sqlite3", "example.db")
if err != nil {
 return err
}
```

#### Insert a new record

```golang

query := lk.Insert("users").
	Set(
		lk.Pair("first_name", "John"),
		lk.Pair("last_name", "Doe"),
	)

if _, err := gateway.Exec(query); err != nil {
  return err
}
```

#### Select all records

```golang
query := lk.Select("id", "first_name", "last_name").From("users")
users := []User{}

if err := gateway.Select(&users, query); err != nil {
  return err
}
```

#### Select a record

```golang
query := lk.Select("id", "first_name", "last_name").
	From("users").
	Where(oak.Condition("first_name").Equal("John"))

user := User{}

if err := gateway.SelectOne(&user, query); err != nil {
  return err
}
```

### SQL Scripts and Commands

Also, it provides a way to work with embeddable SQL scripts by exposing them as
SQL Commands. First of all you have create a script that contains your SQL
statements.

The easies way to generate a SQL script with correct format is by using `oak`
command line interface:

```console
$ oak script create show-sqlite-master
```

The command above will generate a script in your `$PWD/database/script`;

```console
$ tree database/

database/
└── script
    └── 20180328184257.sql
```

It has the following contents:

```sql
-- Auto-generated at Wed Mar 28 18:42:57 CEST 2018
-- name: show-sqlite-master
SELECT type,name,rootpage FROM sqlite_master;
```

The `-- name: show-sqlite-master` comment define the name of the command in
your SQL script. The SQL statement afterwards is considered as the command
body. Note that the command must have only one statement.

Then you can use the `oak` command line interface to execute the command:

```console
$ oak script run show-sqlite-master

Running command 'show-sqlite-master' from '$PWD/database/script'
+-------+-------------------------------+----------+
| TYPE  |             NAME              | ROOTPAGE |
+-------+-------------------------------+----------+
| table | migrations                    |        2 |
| index | sqlite_autoindex_migrations_1 |        3 |
+-------+-------------------------------+----------+
```

You can run the command by using the `Gateway API` as well:

Let's first load the file:

```golang
if err = oak.LoadSQLCommandsFromReader(file); err != nil {
	log.WithError(err).Fatal("Failed to load script")
}
```

Then you can execute the desired script by just passing its name:

```golang
_, err = gateway.Exec(oak.Command("show-sqlite-master"))
```

If you want to run Raw SQL Scripts from your code, you should follow this
example:

```golang
rows, err := gateway.Query(oak.SQL("SELECT * FROM users WHERE id = ?", 5432))
```

### Command Line Interface Advance Usage

By default the CLI work with `sqlite3` database called `oak.db` at your current
directory.

oak supports:

- PostgreSQL
- MySQL
- SQLite

If you want to change the default connection, you can pass it via command line
argument:

```bash
$ oak --database-url [driver-name]://[connection-string] [command]
```

oak uses a URL schema to determines the right database driver. If you want to
pass the connection string via environment variable, you should export
`OAK_DB_URL`.

### Example

You can check our [Getting Started Example](/example).

For more information, how you can change the default behavior you can read the
help documentation by executing:

```bash
$ oak -h
```

## Contributing

We are welcome to any contributions. Just fork the
[project](https://github.com/phogolabs/oak).

*logo made by [Free Pik][logo-author-url]*

[report-img]: https://goreportcard.com/badge/github.com/phogolabs/oak
[report-url]: https://goreportcard.com/report/github.com/phogolabs/oak
[logo-author-url]: https://www.freepik.com/free-photos-vectors/tree
[logo-license]: http://creativecommons.org/licenses/by/3.0/
[oak-url]: https://github.com/phogolabs/oak
[oak-img]: doc/img/logo.png
[codecov-url]: https://codecov.io/gh/phogolabs/oak
[codecov-img]: https://codecov.io/gh/phogolabs/oak/branch/master/graph/badge.svg
[travis-img]: https://travis-ci.org/phogolabs/oak.svg?branch=master
[travis-url]: https://travis-ci.org/phogolabs/oak
[oak-url]: https://github.com/phogolabs/oak
[godoc-url]: https://godoc.org/github.com/phogolabs/oak
[godoc-img]: https://godoc.org/github.com/phogolabs/oak?status.svg
[license-img]: https://img.shields.io/badge/license-MIT-blue.svg
[software-license-url]: LICENSE
[loukoum-url]: https://github.com/ulule/loukoum
[sqlx-url]: https://github.com/jmoiron/sqlx
