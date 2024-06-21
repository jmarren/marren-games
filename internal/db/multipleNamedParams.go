package db

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	"reflect"
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
			Tag:  reflect.StructTag(`xml:"` + CapitalizeFirstLetter(col.Name()) + `"`),
		})
	}
	structFields = append(structFields, reflect.StructField{
		Name: "XMLName",
		Type: reflect.TypeOf(xml.Name{}),
		Tag:  reflect.StructTag(`xml:"answers"`),
	})

	fmt.Println("structFields: ", structFields)

	rowSlice := reflect.MakeSlice(reflect.SliceOf(reflect.StructOf(structFields)), 0, 32)

	fmt.Println("rowSlice: ", rowSlice)

	for outputRows.Next() {
		valPtrs := make([]interface{}, len(cols))
		v := reflect.New(rowSlice.Type().Elem()).Elem()
		for i, col := range cols {
			valPtrs[i] = v.FieldByName(CapitalizeFirstLetter(col.Name())).Addr().Interface()
		}
		err := outputRows.Scan(valPtrs...)
		if err != nil {
			fmt.Println(err)
			return "Error Scanning output into vals", err
		}
		rowSlice = reflect.Append(rowSlice, v)
	}

	fmt.Println("rowSlice: ", rowSlice)
	type StringValue struct {
		String  string `json:"string"`
		isValid bool   `json:"-"`
	}

	type Int64Value struct {
		Int64   int64 `json:"float64"`
		isValid bool  `json:"-"`
	}

	type Answers struct {
		XMLName           xml.Name    `xml:"answers"`
		Answerer_username StringValue `xml:"answerer_username"`
		Answerer_id       Int64Value
		Question_id       Int64Value
		Answer_text       StringValue
	}

	fmt.Println("rowSlice.Index(0): ", rowSlice.Index(0))

	var xmlOutputSlice []string

	for i := 0; i < rowSlice.Len(); i++ {
		fmt.Println("rowSlice.Index(i): ", rowSlice.Index(i))
		xmlOutput, err := xml.Marshal(rowSlice.Index(i).Interface())
		if err != nil {
			fmt.Println("Error marshalling into xml: ", err)
			return "Error Marshalling output into xml", err
		}
		xmlOutputSlice = append(xmlOutputSlice, string(xmlOutput))
	}

	fmt.Println("xmlOutputSlice: ", xmlOutputSlice)
	/*
		// var xmlAnswers Answers
		var xmlAnswers []Answers

		// var xmlOutput []interface{}

		for _, xmlOutput := range xmlOutputSlice {
			var xmlOut Answers
			xmlOutputString := string(xmlOutput)
			fmt.Println("xmlOutputString: ", xmlOutputString)
			xmlErr := xml.Unmarshal(xmlOutput, &xmlOut)
			if xmlErr != nil {
				fmt.Println("xmlErr: ", xmlErr)
				return "Error Unmarshalling output into xml", xmlErr
			}
			xmlAnswers = append(xmlAnswers, xmlOut)
		}

		fmt.Println("xmlAnswers", xmlAnswers)
	*/

	return "testing", nil

	/*

		outputStruct := reflect.StructOf(structFields)

		fmt.Println("outputStruct: ", outputStruct)

		v := reflect.New(outputStruct).Elem()

		fmt.Println("v: ", v)

		outputSlice := reflect.MakeSlice(reflect.SliceOf(outputStruct), 0, 32)

		structFromRowFields := []reflect.StructField{}

		structFromRowField := reflect.StructField{
			Name: "Root",
			Type: outputSlice.Type(),
			Tag:  `xml:"root"`,
		}
		structFromRowFields = append(structFromRowFields, structFromRowField)

		RootStruct := reflect.StructOf(structFromRowFields) // Returns Type

		RSVal := reflect.ValueOf(&RootStruct) // returns value

		fmt.Println("RSVal: ", RSVal)

		RootStructValue := reflect.ValueOf(RootStruct) // returns value

		fmt.Println("RootStructValue: ", RootStructValue)

		RootStructField := RSVal.Index(1) //

		// RootStructField, ok := RootStruct.FieldByName("Root")
		// if !ok {
		// 	fmt.Println("not ok")
		// }

		fmt.Println("RootStructField: ", RootStructField)

		RootStructFieldPtr := &RootStructField

		fmt.Println("RootStructFieldVal: ", RootStructFieldPtr)

		inner := RootStructValue.Elem().FieldByName("Root")

		fmt.Println("inner: ", inner)
	*/

	/*
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

			// fmt.Println(field)
			// inner := append(inner, v)
			// newSlice = append(RootStructField.Index(0), v)

			// RootStructField = reflect.Append(RootStructFieldPtr, v)
			// outputSlice = reflect.Append(outputSlice, v)
		}
	*/
	/*
		fmt.Println("inner: ", inner)

		// RSVal.Set(outputSlice)

		// fmt.Println("type of outputSlice: ", outputSlice.Type())
		//
		// RSSlice, ok := RootStruct.FieldByName("Root").Addr()
		// if !ok {
		// 	fmt.Println("not ok ?? ")
		// }

		// RSVal.FieldByName("Root").Elem().Set(reflect.ValueOf(outputSlice))

		// row1 := outputSlice.Index(0).Interface()

		//
		// fmt.Println("structFromRow: ", structFromRow)
		//
		// fmt.Println("typeof row1: ", reflect.TypeOf(row1))
		//
		xmlOutput, err := xml.Marshal(RootStruct)
		if err != nil {
			fmt.Println("Error marshalling into xml: ", err)
			return "Error Marshalling output into xml", err
		}

		jsonOutput, err := json.Marshal(RootStruct)
		if err != nil {
			fmt.Println(err)
			return "Error Marshalling output into json", err
		}

		fmt.Println("jsonOutput: ", string(jsonOutput))
		fmt.Println("xmlOutput: ", string(xmlOutput))

		// dec := json.NewDecoder(strings.NewReader(jsonOutput))

		type StringValue struct {
			String  string `json:"string"`
			isValid bool   `json:"-"`
		}

		type Int64Value struct {
			Int64   int64 `json:"float64"`
			isValid bool  `json:"-"`
		}

		type Answers struct {
			Answerer_username StringValue `xml:"answerer_username"`
			Answerer_id       Int64Value
			Question_id       Int64Value
			Answer_text       StringValue
		}

		var m Answers
		var xmlAnswers Answers

		unmarshalErr := json.Unmarshal(jsonOutput, &m)
		if unmarshalErr != nil {
			fmt.Println(unmarshalErr)
			return "Error Unmarshalling output", unmarshalErr
		}

		xmlErr := xml.Unmarshal(xmlOutput, &xmlAnswers)
		if xmlErr != nil {
			fmt.Println("xmlErr: ", xmlErr)
			return "Error Unmarshalling output into xml", xmlErr
		}

		// for {
			// 	if err := dec.Decode(&m); err == io.EOF {
			// 		fmt.Println(err)
			// 		break
			// 	} else if err != nil {
			// 		log.Fatal(err)
			// 	}
			// }

			fmt.Println("**************** m ****************")
			fmt.Println("------ ", m)
			fmt.Println("**************** xmlAnswers ****************")
			fmt.Println("------ ", xmlAnswers)

			fmt.Println("m.Answerer_username:", m.Answerer_username)

			fmt.Println("typeof m:", reflect.TypeOf(m))

			// fmt.Println("m.Username.String: ", m.)

			// fmt.Println(row1.FieldByName("Answerer_username").Elem())

			// fmt.Println("outputSlice[1]: ", outputSlice[1])
			//
			// fmt.Printf("outputSlice[1]: %v\n", outputSlice[1])

			// for i := 0; i < len(outputArr); i++ {
			// 	fmt.Println(outputArr[i].answerer_id)
			// }

			// fmt.Println("outputArr: ", outputArr)

			return "", nil
	*/
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
