package db

import (
	"database/sql"
	"fmt"
	"reflect"
	"unicode"
)

func QueryWithMultipleNamedParams(query string, params []sql.NamedArg, anonStruct interface{}) (interface{}, string, error) {
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

	fmt.Println("rows: ", rows)

	// Get the type of the passed-in struct
	dataType := reflect.TypeOf(anonStruct)
	fmt.Println("dataType: ", dataType)

	// Ensure the passed in data is a struct
	if dataType.Kind() != reflect.Struct {
		fmt.Println("expected struct")
	}

	// Create a new instance of the struct type
	structValue := reflect.New(dataType).Elem()

	// Get the number of fields in the struct
	numFields := structValue.NumField()

	// Create a slice to hold the pointers
	slicePointers := make([]interface{}, numFields)

	results := reflect.MakeSlice(reflect.TypeOf(slicePointers), 0, 0)

	for rows.Next() {
		// Create a new instance of the struct type
		newStructPtr := reflect.New(dataType)

		slicePointers := make([]interface{}, numFields)

		for i := 0; i < numFields; i++ {
			slicePointers[i] = newStructPtr.Elem().Field(i).Addr().Interface()
		}

		err := rows.Scan(slicePointers...)
		if err != nil {
			fmt.Println("error scanning rows into struct: ", err)
			return nil, "Error scanning rows into struct", err
		}

		results = reflect.Append(results, newStructPtr)
	}

	fmt.Println("results: ", results)

	return results.Interface(), "Record created successfully", nil
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
