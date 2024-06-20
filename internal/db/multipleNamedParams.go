package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"reflect"
	"strings"
	"unicode"
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
	cols, err := outputRows.ColumnTypes()
	if err != nil {
		fmt.Println(err)
		return "Error getting cols from output ", err
	}

	fmt.Println("cols", cols)
	structFields := []reflect.StructField{}

	for _, col := range cols {
		structFields = append(structFields, reflect.StructField{
			Name: CapitalizeFirstLetter(col.Name()),
			Type: col.ScanType(),
			Tag:  reflect.StructTag(col.Name()),
		})
	}

	outputStruct := reflect.StructOf(structFields)

	fmt.Println("outputStruct: ", outputStruct)

	v := reflect.New(outputStruct).Elem()

	fmt.Println("v: ", v)

	// outputSlice := []interface{}{}
	outputSlice := reflect.MakeSlice(reflect.SliceOf(outputStruct), 0, 32)

	for outputRows.Next() {
		valPtrs := make([]interface{}, len(cols))

		v := reflect.New(outputStruct).Elem()
		for i, col := range cols {
			valPtrs[i] = v.FieldByName(CapitalizeFirstLetter(col.Name())).Addr().Interface()
			fmt.Println(v.FieldByName(CapitalizeFirstLetter(col.Name())).Addr().Interface())
		}
		err := outputRows.Scan(valPtrs...)
		if err != nil {
			fmt.Println(err)
			return "Error Scanning output into vals", err
		}
		outputSlice = reflect.Append(outputSlice, v)
	}

	row1 := outputSlice.Index(0).Interface()

	jsonOutput, err := json.Marshal(row1)
	if err != nil {
		fmt.Println(err)
		return "Error Marshalling output", err
	}

	fmt.Println("jsonOutput: ", string(jsonOutput))

	dec := json.NewDecoder(strings.NewReader(string(jsonOutput)))

	type StringValue struct {
		String string
		Valid  bool
	}

	type Int64Value struct {
		Int64 int
		Valid bool
	}

	type Data struct {
		Answerer_username StringValue
		Answerer_id       Int64Value
		Question_id       Int64Value
		Answer_text       StringValue
	}

	var m Data

	for {
		if err := dec.Decode(&m); err == io.EOF {
			fmt.Println(err)
			break
		} else if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("m -----> \n", m)

	fmt.Println("m.Answerer_username.String:", m.Answerer_username.String)

	// fmt.Println(row1.FieldByName("Answerer_username").Elem())

	// fmt.Println("outputSlice[1]: ", outputSlice[1])
	//
	// fmt.Printf("outputSlice[1]: %v\n", outputSlice[1])

	// for i := 0; i < len(outputArr); i++ {
	// 	fmt.Println(outputArr[i].answerer_id)
	// }

	// fmt.Println("outputArr: ", outputArr)

	/*

		for outputRows.Next() {
			outputArr := reflect.New(outputStructArray).Elem()

			fmt.Println("outputArr: ", outputArr)

			// outputType := reflect.TypeOf(reflect.StructOf(structFields))
			//
			// fmt.Println(outputType)

			// outputValsStruct := []reflect.TypeOf(outputType){}

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
	*/
	return "", nil
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
