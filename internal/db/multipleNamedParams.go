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

func QueryWithMultipleNamedParams(query string, params []sql.NamedArg, createNewSlice func() RowContainer, typ reflect.Type) (interface{}, string, error) {
	// Convert Named Params to Interface so they can be passed to Query
	var paramsInterface []interface{}
	for _, param := range params {
		paramsInterface = append(paramsInterface, param)
	}

	fmt.Println("Query:  ", query)

	// Execute Query
	rows, err := Sqlite.Query(query, paramsInterface...)
	if err != nil {
		fmt.Println(err)
		return nil, "Error Executing Query", err
	}

	fmt.Println("rows: ", rows)

	// results := []interface{}{}
	//
	// results := []interface{}{}

	results := reflect.MakeSlice(reflect.SliceOf(typ), 0, 0)
	fmt.Println("results:", results)

	for rows.Next() {

		newSlice := createNewSlice()

		slicePtrs := newSlice.GetPtrs()

		fmt.Println("slicePtrs: ", slicePtrs)

		err := rows.Scan(slicePtrs...)
		fmt.Println("newSlice: ", newSlice)

		if err != nil {
			return nil, "error scanning rows", err
		}

		// Using reflection to append to the results slice
		resultValue := reflect.ValueOf(newSlice)
		if resultValue.Type().ConvertibleTo(typ) {
			concreteType := resultValue.Convert(typ)
			results = reflect.Append(results, concreteType)
		} else {
			fmt.Println("Unexpected type")
		}

	}

	fmt.Println("results: ", results)

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
