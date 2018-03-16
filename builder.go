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
	And       = loukoum.And
	Column    = loukoum.Column
	Condition = loukoum.Condition
	Delete    = loukoum.Delete
	DoNothing = loukoum.DoNothing
	DoUpdate  = loukoum.DoUpdate
	Insert    = loukoum.Insert
	On        = loukoum.On
	Or        = loukoum.Or
	Order     = loukoum.Order
	Pair      = loukoum.Pair
	Raw       = loukoum.Raw
	Select    = loukoum.Select
	Table     = loukoum.Table
	Update    = loukoum.Update
)
