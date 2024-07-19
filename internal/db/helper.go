package db

import "database/sql"

type Val struct {
	Name  string
	Value any
}

func SqlName(vals []Val) []sql.NamedArg {
	var result []sql.NamedArg

	for _, arg := range vals {
		result = append(result, sql.Named(arg.Name, arg.Value))
	}

	return result
}
