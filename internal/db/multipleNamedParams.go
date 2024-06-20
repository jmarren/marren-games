package db

import (
	"database/sql"
	"fmt"
)

func QueryWithMultipleNamedParams(query string, params []sql.NamedArg) (string, error) {
	// Convert Named Params to Interface so they can be passed to Query
	var paramsInterface []interface{}
	for _, param := range params {
		paramsInterface = append(paramsInterface, param)
	}

	// Execute Query
	outputRows, err := Sqlite.Query(query, paramsInterface...)
	if err != nil {
		fmt.Println(err)
		return "Error Executing Query", err
	}

	// Get Columns
	cols, err := outputRows.Columns()
	if err != nil {
		fmt.Println(err)
		return "Error getting cols from output ", err
	}
	// Create vals and valPtrs to store output
	colLen := len(cols)
	vals := make([]interface{}, colLen)
	valPtrs := make([]interface{}, colLen)

	response := ""

	// Loop through output and store in vals
	for outputRows.Next() {
		for i := range cols {
			valPtrs[i] = &vals[i]
		}
		err := outputRows.Scan(valPtrs...)
		if err != nil {
			fmt.Println(err)
			return "Error Scanning output into vals", err
		}
		for i, col := range cols {
			val := vals[i]

			b, ok := val.([]byte)
			var v interface{}
			if ok {
				v = string(b)
			} else {
				v = val
			}
			fmt.Println(col, v)
			response = fmt.Sprintf("%v\n %v: %v", response, col, v)
		}
		response = fmt.Sprintf("%v\n-----------", response)
	}

	return response, nil
}
