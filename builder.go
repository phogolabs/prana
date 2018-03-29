package gom

import (
	"github.com/ulule/loukoum"
)

const (
	// InnerJoin is used for "INNER JOIN" in join statement.
	InnerJoin = loukoum.InnerJoin
	// LeftJoin is used for "LEFT JOIN" in join statement.
	LeftJoin = loukoum.LeftJoin
	// RightJoin is used for "RIGHT JOIN" in join statement.
	RightJoin = loukoum.RightJoin
	// Asc is used for "ORDER BY" statement.
	Asc = loukoum.Asc
	// Desc is used for "ORDER BY" statement.
	Desc = loukoum.Desc
)

// Map is a key/value map.
type Map = loukoum.Map

var (
	// And is a wrapper to create a new InfixExpression statement.
	And = loukoum.And
	// Column is a wrapper to create a new Column statement.
	Column = loukoum.Column
	// Condition is a wrapper to create a new Identifier statement.
	Condition = loukoum.Condition
	// Delete starts a DeleteBuilder using the given table as from clause.
	Delete = loukoum.Delete
	// DoNothing is a wrapper to create a new ConflictNoAction statement.
	DoNothing = loukoum.DoNothing
	// DoUpdate is a wrapper to create a new ConflictUpdateAction statement.
	DoUpdate = loukoum.DoUpdate
	// Insert starts an InsertBuilder using the given table as into clause.
	Insert = loukoum.Insert
	// On is a wrapper to create a new On statement.
	On = loukoum.On
	// Or is a wrapper to create a new InfixExpression statement.
	Or = loukoum.Or
	// Order is a wrapper to create a new Order statement.
	Order = loukoum.Order
	// Pair takes a key and its related value and returns a Pair.
	Pair = loukoum.Pair
	// Raw is a wrapper to create a new Raw expression.
	Raw = loukoum.Raw
	// Select starts a SelectBuilder using the given columns.
	Select = loukoum.Select
	// Table is a wrapper to create a new Table statement.
	Table = loukoum.Table
	// Update starts an Update builder using the given table.
	Update = loukoum.Update
)
