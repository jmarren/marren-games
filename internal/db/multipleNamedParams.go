package db

import (
	"database/sql"
	"fmt"
	"reflect"
	"unicode"
)

type RowContainer interface {
	GetPtrs() []interface{}
}

func QueryWithMultipleNamedParams(query string, params []sql.NamedArg, concreteType reflect.Type) (interface{}, string, error) {
	// Convert Named Params to Interface so they can be passed to Query
	var paramsInterface []interface{}
	for _, param := range params {
		paramsInterface = append(paramsInterface, param)
	}
	// Execute Query
	rows, err := Sqlite.Query(query, paramsInterface...)
	if err != nil {
		fmt.Println(err)
		return nil, "Error Executing Query", err
	}

	structValue := reflect.New(concreteType).Elem()

	// Create a slice to hold pointers
	// to each field
	slicePtrs := make([]interface{}, structValue.NumField())

	// Create a slice to hold slicePtrs
	// for each row
	results := reflect.MakeSlice(reflect.TypeOf(slicePtrs), 0, 0)

	for rows.Next() {

		newStructPtr := reflect.New(concreteType)

		// Create a slice to hold pointers to each field
		slicePtrs := make([]interface{}, concreteType.NumField())

		// Iterate over the fields and get pointers
		for i := 0; i < concreteType.NumField(); i++ {
			field := newStructPtr.Elem().Field(i)
			slicePtrs[i] = field.Addr().Interface()
		}

		// Scan the rows into the slicePtrs
		err := rows.Scan(slicePtrs...)
		if err != nil {
			return nil, "error scanning rows", err
		}

		// Append the newStructPtr to the restuls slice
		results = reflect.Append(results, newStructPtr)
	}

	return results.Interface(), "successfully executed query", nil
}

func ExecTestWithNamedParams(query string, params []sql.NamedArg) (string, error) {
	var paramsInterface []interface{}
	for _, param := range params {
		paramsInterface = append(paramsInterface, param)
	}

	response, err := Sqlite.Exec(query, paramsInterface...)
	if err != nil {
		fmt.Println(err)
		return "Error Executing Exec Query", err
	}

	fmt.Println("response: ", response)

	return "Record created successfully", nil
}

// Capitalize the first letter of a string
func CapitalizeFirstLetter(s string) string {
	// Check if the string is empty
	if len(s) == 0 {
		return s
	}

	// Convert the string to a rune slice for proper handling of UTF-8 characters
	runes := []rune(s)

	// Capitalize the first rune
	runes[0] = unicode.ToUpper(runes[0])

	// Convert the rune slice back to a string
	return string(runes)
}
