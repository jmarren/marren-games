package db

import (
	"database/sql"
	"fmt"
	"reflect"
	"unicode"
)

// // Function to get an array of pointers to each field of a struct
//
//	func getFieldPointers(s interface{}) []any {
//		// Get the reflect.Value of the struct
//		v := reflect.ValueOf(s)
//		fmt.Println("v: ", v)
//		if v.Kind() != reflect.Struct {
//			panic("expected a struct")
//		}
//
//		// Create a slice to hold the pointers
//		pointers := reflect.MakeSlice(, len int, cap int)
//
//		// Iterate over the fields and get pointers to each field
//		for i := 0; i < v.NumField(); i++ {
//			field := v.Field(i)
//			// Create a new interface{} to hold the pointer
//			var ptr interface{}
//			// Set the pointer to the address of the field
//			reflect.ValueOf(&ptr).Elem().Set(field.Addr())
//			// Add the pointer to the slice
//			pointers[i] = &ptr
//		}
//
//		return pointers
//	}
//
// // Function to create a new instance of an anonymous struct type
//
//	func createNewStructInstance(data interface{}) interface{} {
//		// Get the type of the passed-in struct
//		dataType := reflect.TypeOf(data)
//
//		// Ensure the passed-in data is a struct
//		if dataType.Kind() != reflect.Struct {
//			fmt.Println("Passed data is not a struct")
//			return nil
//		}
//
//		// Create a new instance of the struct type
//		newInstance := reflect.New(dataType).Elem()
//		fmt.Println("newInstance: ", newInstance)
//
//		// Return the new instance as an interface{}
//		return newInstance.Interface()
//	}
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
