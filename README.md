# Prana

[![Documentation][godoc-img]][godoc-url]
![License][license-img]
[![Build Status][travis-img]][travis-url]
[![Coverage][codecov-img]][codecov-url]
[![Go Report Card][report-img]][report-url]

*Golang Database Manager*

[![Prana][prana-img]][prana-url]

## Overview

Prana is a package for rapid application development with relational databases in
Golang.  It has a command line interface that provides:

- SQL Migrations
- Embedded SQL Scripts
- Model generation from SQL schema

## Installation

#### GitHub

```console
$ go get -u github.com/phogolabs/prana
$ go install github.com/phogolabs/prana/cmd/prana
```

#### Homebrew (for Mac OS X)

```console
$ brew tap phogolabs/tap
$ brew install prana
```

## Introduction

Note that we may introduce breaking changes until we reach v1.0.

### SQL Migrations

Each migration is a SQL script that contains two operations for upgrade and
rollback. They are labelled with the following comments:

- `-- name: up` for upgrade
- `-- name: down` for revert

In order to prepare the project for migration, you have to set it up:

```console
$ prana migration setup
```

Then you can create a migration with the following command:

```console
$ prana migration create schema
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
$ prana migration run
```

If you want to rollback the migration you have to revert it:

```console
$ prana migration revert
```

If you have an SQL script that is compatible with particular database, you can
append the database's driver name suffix. For instance if you want to run part
of a particular migration for MySQL, you should have the following directory
tree:

```
$ tree database

database/
└── migration
    ├── 00060524000000_setup.sql
    ├── 20180406190015_users.sql
    ├── 20180406190015_users_mysql.sql
    ├── 20180406190015_users_postgres.sql
    └── 20180406190015_users_sqlite3.sql
```

Prana will execute the following migrations with `users` suffix, when MySQL
driver is used:

- `20180406190015_users.sql`
- `20180406190015_users_mysql.sql`

## SQL Schema and Code Generation

Let's assume that we want to generate a mode for the `users` table.

You can use the `prana` command line interface to generate a package that
contains Golang structs, which represents each table from the desired schema.

For that purpose you should call the following subcommand:

```bash
$ prana model sync
```

By default the command will place the generated code in single `schema.go` file
in `$PWD/database/model` package for the default database schema. Any other
schemas will be placed in the same package but in separate files. You can
control the behavior by passing `--keep-schema` flag which will
cause each schema to be generated in own package under the
`/$PWD/database/model` package.

You can print the source code without generating a package by executing the
following command:

```bash
$ prana model print
```

Note that you can specify the desired schema or tables by providing the correct
arguments.

If you pass `--extra-tag` argument, you can specify which tag to be included in
your final result. Supported extra tags are:

- [json](https://golang.org/pkg/encoding/json/)
- [xml](https://golang.org/pkg/encoding/xml/)
- [validate](https://github.com/go-playground/validator/blob/v9/_examples/simple/main.go#L11) to validates fields by [validator](https://github.com/go-playground/validator) package

The model representation of the users table is:

```golang
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
[prana.Gateway](https://godoc.org/github.com/phogolabs/prana#Gateway) as well as
[sqlx](https://github.com/jmoiron/sqlx).

If you wan to generate models for [gorm](http://gorm.io), you should
pass `--orm-tag gorm`. Note that constraints like unique or indexes are not
included for now.

```console
$ prana model --orm-tag gorm -e json -e xml -e validate sync
```

The command above will produce the following model:

```golang
package model

import null "gopkg.in/volatiletech/null.v6"

// User represents a data base table 'users'
type User struct {
	// ID represents a database column 'id' of type 'INT PRIMARY KEY NOT NULL'
	ID int `gorm:"column:id;type:int;primary_key;not null" json:"id" xml:"id" validate:"required"`

	// FirstName represents a database column 'first_name' of type 'TEXT NOT NULL'
	FirstName string `gorm:"column:first_name;type:text;not null" json:"first_name" xml:"first_name" validate:"required"`

	// LastName represents a database column 'last_name' of type 'TEXT NULL'
	LastName null.String `gorm:"column:last_name;type:text;null" json:"last_name" xml:"last_name" validate:"-"`
}
```

### SQL Scripts and Commands

Also, it provides a way to work with embeddable SQL scripts which can be
executed easily by [OAK][oak-url] as SQL Routines. You can see the
[OAK example](https://github.com/phogolabs/oak/tree/master/example) to
understand more about that. First of all you have create a script that contains
your SQL statements.

The easies way to generate a SQL script with correct format is by using `prana`
command line interface:

```console
$ prana routine create show-sqlite-master
```

The command above will generate a script in your `$PWD/database/routine`;

```console
$ tree database/

database/
└── routine
    └── 20180328184257.sql
```

You can enable the script for particular type of database by adding the driver
name as suffix: `20180328184257_slite3.sql`.

It has the following contents:

```sql
-- Auto-generated at Wed Mar 28 18:42:57 CEST 2018
-- name: show-sqlite-master
SELECT type,name,rootpage FROM sqlite_master;
```

The `-- name: show-sqlite-master` comment define the name of the command in
your SQL script. The SQL statement afterwards is considered as the command
body. Note that the command must have only one statement.

Then you can use the `prana` command line interface to execute the command:

```console
$ prana script run show-sqlite-master

Running command 'show-sqlite-master' from '$PWD/database/script'
+-------+-------------------------------+----------+
| TYPE  |             NAME              | ROOTPAGE |
+-------+-------------------------------+----------+
| table | migrations                    |        2 |
| index | sqlite_autoindex_migrations_1 |        3 |
+-------+-------------------------------+----------+
```

You can also generate all CRUD operations for given table. The command below
will generate a SQL script that contains SQL queries for each table in the
default schema:

```consol
$ prana routine sync
```

It will produce the following script in `$PWD/database/rotuine`:

```sql
-- name: select-all-users
SELECT * FROM users;

-- name: select-user
SELECT * FROM users
WHERE id = ?;

-- name: insert-user
INSERT INTO users (id, first_name, last_name)
VALUES (?, ?, ?);

-- name: update-user
UPDATE users
SET first_name = ?, last_name = ?
WHERE id = ?;

-- name: delete-user
DELETE FROM users
WHERE id = ?;
```

### Command Line Interface Advance Usage

By default the CLI work with `sqlite3` database called `prana.db` at your current
directory.

Prana supports:

- PostgreSQL
- MySQL
- SQLite

If you want to change the default connection, you can pass it via command line
argument:

```bash
$ prana --database-url [driver-name]://[connection-string] [command]
```

prana uses a URL schema to determines the right database driver. If you want to
pass the connection string via environment variable, you should export
`PRANA_DB_URL`.

### Help

For more information, how you can change the default behavior you can read the
help documentation by executing:

```console
$ prana -h

NAME:
   prana - Golang Database Manager

USAGE:
   prana [global options]

VERSION:
   1.0

COMMANDS:
     migration  A group of commands for generating, running, and reverting migrations
     model      A group of commands for generating object model from database schema
     routine    A group of commands for generating, running, and removing SQL commands
     help, h    Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --database-url value  Database URL (default: "sqlite3://prana.db") [$PRANA_DB_URL]
   --log-format value    format of the logs [$PRANA_LOG_FORMAT]
   --log-level value     level of logging (default: "info") [$PRANA_LOG_LEVEL]
   --help, -h            show help
   --version, -v         print the version
```

## Contributing

We are welcome to any contributions. Just fork the
[project](https://github.com/phogolabs/prana).

*logo made by [Free Pik][logo-author-url]*

[report-img]: https://goreportcard.com/badge/github.com/phogolabs/prana
[report-url]: https://goreportcard.com/report/github.com/phogolabs/prana
[logo-author-url]: https://www.freepik.com/free-vector/abstract-cross-logo-template_1185919.htm
[logo-license]: http://creativecommons.org/licenses/by/3.0/
[prana-url]: https://github.com/phogolabs/prana
[prana-img]: doc/img/logo.png
[codecov-url]: https://codecov.io/gh/phogolabs/prana
[codecov-img]: https://codecov.io/gh/phogolabs/prana/branch/master/graph/badge.svg
[travis-img]: https://travis-ci.org/phogolabs/prana.svg?branch=master
[travis-url]: https://travis-ci.org/phogolabs/prana
[prana-url]: https://github.com/phogolabs/prana
[godoc-url]: https://godoc.org/github.com/phogolabs/prana
[godoc-img]: https://godoc.org/github.com/phogolabs/prana?status.svg
[license-img]: https://img.shields.io/badge/license-MIT-blue.svg
[software-license-url]: LICENSE
[loukoum-url]: https://github.com/ulule/loukoum
[oak-url]: https://github.com/phogolabs/oak
[sqlx-url]: https://github.com/jmoiron/sqlx
