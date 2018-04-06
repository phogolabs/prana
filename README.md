# GOM

[![Documentation][godoc-img]][godoc-url]
![License][license-img]
[![Build Status](https://travis-ci.org/phogolabs/gom.svg?branch=master)](https://travis-ci.org/phogolabs/gom)

## Overview

GOM is a package that speed up working with SQL database by providing
command line interface maintaining schema migrations and generating SQL
scripts, as well as facilitating working with [loukoum][loukoum-url].

## Installation

```console
go get -u github.com/phogolabs/gom
go install -u github.com/phogolabs/gom/cmd/gom
```

## Introduction

The GOM Gateway gets advantage by using internally [sqlx][sqlx-url].

### SQL Queries with Loukoum

Before working with loukoum and gom, you should import the desired packages:

```golang
import (
  lk "github.com/ulule/loukoum"
  "github.com/phogolabs/gom"
)
```

Let's assume that we have the following model and database:

```golang
type User struct {
	ID        int64  `db:"id"`
	FirstName string `db:"last_name"`
	LastName  string `db:"first_name"`
}
```

Let's first establish the connection:

```golang
gateway, err := gom.Open("sqlite3", "example.db")
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
	Where(gom.Condition("first_name").Equal("John"))

user := User{}

if err := gateway.SelectOne(&user, query); err != nil {
  return err
}
```

### SQL Scripts and Commands

Also, it provides a way to work with SQL scripts by exposing them as GOM
Commands. First of all you have create a script that contains your SQL
statements.

The easies way to generate a SQL script with correct format is by using `gom`
command line interface:

```console
$ gom script create show-sqlite-master
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
your SQL script. Any SQL statements afterwards are considered as the command
body.

Then you can use the `gom` command line interface to execute the command:

```console
$ gom script run show-sqlite-master

Running command 'show-sqlite-master'
+-------+-------------------------------+----------+
| TYPE  |             NAME              | ROOTPAGE |
+-------+-------------------------------+----------+
| table | migrations                    |        2 |
| index | sqlite_autoindex_migrations_1 |        3 |
+-------+-------------------------------+----------+
Running command 'show-sqlite-master' completed successfully
```

You can run the command by using the `Gateway API` as well:

```golang
err := gom.LoadDir("./database/script")

if err != nil {
	return err
}

cmd := gom.Command("show-sqlite-master")

_, err = gateway.Exec(cmd)
return err
```

### SQL Migrations

The SQL Migration are based on the SQL command approach. Each migration is a
SQL script that contains `up` and `down` commands.

In order to prepare the project for migration, you have set it up:

```console
$ gom migration setup
```

Then you can create a migration with the following command:

```console
$ gom migration create schema
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
$ gom migration run
```

If you want to rollback the migration you have to revert it:

```console
$ gom migration revert
```

## SQL Schema and Code Generation

Let's assume that we want to generate a mode for the `users` table.

You can use the `gom` command line interface to generate a package that
contains Golang structs, which represents each table from the desired schema.

For that purpose you should call the following subcommand:

```bash
$ gom schema sync
```

By default the command will place the generated code in single `model.go` file in
`$PWD/database/model` package for the default database schema.

You can print the source code without generating a package by executing the
following command:

```bash
$ gom schema print
```

Note that you can specify the desired schema or tables by providing the correct
arguments.

The entity representation of the users table is:

```golang
// User represents a data base table 'users' from 'default' schema
type User struct {
	// Id represents a database column 'id' of type 'INT NULL'
	Id null.Int `db:"id" json:"id"`

	// FirstName represents a database column 'first_name' of type 'TEXT NULL'
	FirstName null.String `db:"first_name" json:"first_name"`

	// LastName represents a database column 'last_name' of type 'TEXT NULL'
	LastName null.String `db:"last_name" json:"last_name"`
}
```

### Command Line Interface Advance Usage

By default the CLI work with `sqlite3` database called `gom.db` at your current
directory.

GOM supports:

- PostgreSQL
- MySQL
- SQLite

If you want to change the default connection, you can pass it via command line
argument:

```bash
$ gom --database-url mysql://root@./gom_demo [command]
```

GOM uses a URL schema to determines the right database driver. If you want to
pass the connection string via environment variable, you should export
`GOM_DB_URL`.

### Example

You can check our [Getting Started Example](/example).

For more information, how you can change the default behavior you can read the
help documentation by executing:

```bash
$ gom -h
```

## Contributing

We are welcome to any contributions. Just fork the
[project](https://github.com/phogolabs/gom).

[gom-url]: https://github.com/phogolabs/gom
[godoc-url]: https://godoc.org/github.com/phogolabs/gom
[godoc-img]: https://godoc.org/github.com/phogolabs/gom?status.svg
[license-img]: https://img.shields.io/badge/license-MIT-blue.svg
[software-license-url]: LICENSE
[loukoum-url]: https://github.com/ulule/loukoum
[sqlx-url]: https://github.com/jmoiron/sqlx
