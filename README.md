# GOM

[![Documentation][godoc-img]][godoc-url]
![License][license-img]

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

The `GOM` has aliases to [loukoum][loukoum-url] in order
to simplify your imports. It provides a database wrapper API to work with any
loukoum built queries.

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
query := gom.Insert("users").
	Set(
		gom.Pair("first_name", "John"),
		gom.Pair("last_name", "Doe"),
	)

if _, err := gateway.Exec(query); err != nil {
  return err
}
```

#### Select all records

```golang
query := gom.Select("id", "first_name", "last_name").From("users")
users := []User{}

if err := gateway.Select(&users, query); err != nil {
  return err
}
```

#### Select a record

```golang
query := gom.Select("id", "first_name", "last_name").
	From("users").
	Where(gom.Condition("first_name").Equal("John"))

user := User{}

if err := gateway.SelectOne(&user, query); err != nil {
  return err
}
```

### SQL Commands

Also, it provides a way to work with SQL scripts by exposing them as GOM
Commands. First of all you have create a script that contains your SQL
statements.

The easies way to generate a SQL script with correct format is by using `gom`
command line interface:

```console
$ gom command create show-sqlite-master
```

The command above will generate a script in your `$PWD/database/command`;

```console
$ tree database/

database/
└── command
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
$ gom command run show-sqlite-master

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
err := gom.LoadDir("./database/command")

if err != nil {
	return err
}

cmd := gom.Command("show-sqlite-master")

_, err = gateway.Exec(cmd)
return err
```

### SQL Migrations

The SQL Migration are based on the command approach. But instead, the migration
SQL scripts are created in `/database/migration` directory. Each migration has
`up` and `down` commands.

In order to prepare the project for migration, you have setup it:

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

The `20180329162010_schema.sql` migration has the following formate:

```sql
-- Auto-generated at Thu Mar 29 16:20:10 CEST 2018
-- Please do not change the name attributes

-- name: up
CREATE TABLE users (
  id INT PRIMARY KEY,
  first_name TEXT,
  last_name TEXT
)

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
## Contributing

We are welcome to any contributions. Just fork the
[project](https://github.com/ulule/loukoum)

[gom-url]: https://github.com/phogolabs/gom
[godoc-url]: https://godoc.org/github.com/phogolabs/gom
[godoc-img]: https://godoc.org/github.com/phogolabs/gom?status.svg
[license-img]: https://img.shields.io/badge/license-MIT-blue.svg
[software-license-url]: LICENSE
[loukoum-url]: https://github.com/ulule/loukoum
[sqlx-url]: https://github.com/jmoiron/sqlx
