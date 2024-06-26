package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
)

func DynamicQuery(query string, params []sql.NamedArg, anonStruct interface{}) (reflect.Value, string, error) {
	// Convert Named Params to Interface so they can be passed to Query
	var paramsInterface []interface{}
	for _, param := range params {
		paramsInterface = append(paramsInterface, param)
	}
	// Execute Query
	rows, err := Sqlite.Query(query, paramsInterface...)
	if err != nil {
		fmt.Println(err)
		return reflect.Value{}, "Error Executing Query", err
	}

	// Craete a flattened version of anonStruct
	flattenedStructType, jsonOutputs, regularOutputs := CreateFlatStructType(anonStruct)
	numFields := flattenedStructType.NumField()

	// Create a slice to hold the pointers
	// pointers := make([]interface{}, numFields)
	// anonStructSliceType := reflect.SliceOf(reflect.TypeOf(anonStruct))
	// results := reflect.MakeSlice(reflect.TypeOf(pointers), 0, 0)

	newStructPtr := reflect.New(flattenedStructType).Elem()

	for rows.Next() {
		// Create a new instance of the struct type

		slicePointers := make([]interface{}, numFields)

		for i := 0; i < numFields; i++ {
			slicePointers[i] = newStructPtr.Field(i).Addr().Interface()
		}

		err := rows.Scan(slicePointers...)
		if err != nil {
			fmt.Println("error scanning rows into struct: ", err)
			return reflect.Value{}, "Error scanning rows into struct", err
		}

		if err != nil {
			panic("error while unmarshaling json outputs")
		}
	}
	s, err := UnmarshalIntoType(anonStruct, newStructPtr, jsonOutputs, regularOutputs)
	if err != nil {
		panic(err)
	}

	finalDataType := reflect.TypeOf(anonStruct)
	dereferencedS := s.Elem()

	var final reflect.Value

	if dereferencedS.Type().ConvertibleTo(finalDataType) {
		fmt.Println(" &&&&&  Yes it is convertible &&&&&&&&&")
		final = dereferencedS.Convert(finalDataType)
	} else {
		fmt.Println(" &&&&&  not convertible &&&&&&&&&")
	}

	finalFinal := reflect.ValueOf(final).Interface()

	fmt.Println("------------------- Final -------------- ")
	fmt.Println("----------------- s: ", s, "------------------")
	fmt.Println("----------------- final: ", final, "------------------")
	fmt.Println("----------------- reflect.TypeOf(final): ", reflect.TypeOf(final), "------------------")
	fmt.Println("----------------- finalFinal: ", finalFinal, "------------------")
	fmt.Println("----------------- reflect.TypeOf(finalFinal): ", reflect.TypeOf(finalFinal), "------------------")
	fmt.Println("------------------------------------------")

	fmt.Print("\n\n ^^^^^^^^^^^^^^^^^")
	// result, err := UnmarshalIntoType(anonStruct, newStructPtr, jsonOutputs)

	return final, "success", nil
}

func UnmarshalIntoType(anonStruct interface{}, newStructPtr reflect.Value, jsonOutputs []FieldInfo, regularOutputs []FieldInfo) (reflect.Value, error) {
	s, err := CreateStructPtrOfSameType(anonStruct)
	if err != nil {
		return reflect.Value{}, errors.New("error creating pointer to new anonStruct")
	}

	for _, v := range regularOutputs {
		err := updateField(s, v.Name, newStructPtr.FieldByName(v.Name).Interface())
		if err != nil {
			panic("error from updateField: " + err.Error())
		}
	}

	for _, v := range jsonOutputs {
		jsonString := newStructPtr.FieldByName(v.Name).FieldByName("String").Interface()

		// Create a new instance of the slice type
		container := reflect.New(v.Type).Elem()

		// Since containerValue is already the correct type, assign it directly to concrete
		concrete := container.Addr().Interface()

		fmt.Println("concrete: ", concrete)

		jsonAsserted, ok := jsonString.(string)
		if !ok {
			panic("not a valid json string")
		}

		dec := json.NewDecoder(strings.NewReader(jsonAsserted))

		if err := dec.Decode(concrete); err != nil && err != io.EOF {
			fmt.Println(err)
			panic("Error while decoding json" + err.Error())
		}

		err := updateField(s, v.Name, concrete)
		if err != nil {
			panic("error from updateField: " + err.Error())
		}

		fmt.Println("concrete: ", concrete)

	}

	// if results.Kind() == reflect.Slice {
	// 		for i := 0; i < results.Len(); i++ {
	// 			item := results.Index(i).Interface()
	//
	// 			// dereference the pointer to get the underlying struct for each slice item
	// 			dereferencedItem := reflect.Indirect(reflect.ValueOf(item)).Interface()
	//
	// 			// Convert the dereferencedItem to the concrete type specified in routeConfig
	// 			dereferencedItemValue := reflect.ValueOf(dereferencedItem)
	//
	// 			if dereferencedItemValue.Type().ConvertibleTo(dataType) {
	// 				concrete := reflect.ValueOf(dereferencedItem).Convert(dataType)
	// 				concreteDataSlice = reflect.Append(concreteDataSlice, concrete)
	//
	// 			} else {
	// 				fmt.Println("Unexpected type")
	// 			}
	// 		}
	// 	} else {
	// 		fmt.Println("Unexpected result type")
	// 	}
	//
	// 	return results.Interface(), "success", ni

	// finalDataType := reflect.TypeOf(anonStruct)
	// dereferencedS := reflect.Indirect(reflect.ValueOf(s))
	// dereferencedSValue := reflect.ValueOf(dereferencedS)
	// var final interface{}
	//
	//
	// if dereferencedSValue.Type().ConvertibleTo(finalDataType) {
	// 	final = reflect.ValueOf(dereferencedS).Convert(finalDataType)
	// }
	//
	// fmt.Println("------------------- Final -------------- ")
	// fmt.Println("----------------- s: ", s, "------------------")
	// fmt.Println("----------------- final: ", final, "------------------")
	// fmt.Println("------------------------------------------")

	return s, nil
}

func updateField(s reflect.Value, fieldName string, value interface{}) error {
	// Get the value of the struct
	// structValue := reflect.ValueOf(s)

	fmt.Print("\n\n")
	fmt.Println(s)
	fmt.Println("reflect.typeof(s): ", reflect.TypeOf(s))
	fmt.Println("value provided: ", value)
	fmt.Println("reflect.TypeOf(value): ", reflect.TypeOf(value))
	fmt.Print("\n\n")

	field := s.Elem().FieldByName(fieldName)

	fmt.Println("** field: ", field)

	// Check if the field exists and can be set
	if !field.IsValid() {
		return fmt.Errorf("no such field: %s in struct", fieldName)
	}
	if !field.CanSet() {
		return fmt.Errorf("cannot set field: %s", fieldName)
	}

	// Get the value to be set
	valueToSet := reflect.ValueOf(value)

	if valueToSet.Kind() != reflect.Struct {
		valueToSet = valueToSet.Elem()
	}

	fmt.Println("** valueToSet: ", valueToSet)

	fmt.Printf("\n\n Trying to set field of type %v to value of type %v ", field.Type(), valueToSet.Type())

	// Check if the type of the value matches the field type
	if field.Type() != valueToSet.Type() {
		return fmt.Errorf("provided value type doesn't match struct field type")
	}

	// Set the field
	field.Set(valueToSet)
	return nil
}

func CreateStructPtrOfSameType(s interface{}) (reflect.Value, error) {
	dataType := reflect.TypeOf(s)
	if dataType.Kind() != reflect.Struct {
		fmt.Println("Error in CreateStructOfSameType: expected struct but got ", dataType.Kind())
		return reflect.Value{}, errors.New("error: expected struct")
	}
	newS := reflect.New(dataType)
	return newS, nil
}

func CreateFlatStructType(s interface{}) (reflect.Type, []FieldInfo, []FieldInfo) {
	v := reflect.ValueOf(s)
	t := reflect.TypeOf(s)
	var newStructFields []reflect.StructField

	var jsonFields []FieldInfo
	var regularFields []FieldInfo
	for i := 0; i < t.NumField(); i++ {
		if isSQLType(t.Field(i).Type) {
			newStructFields = append(newStructFields,
				reflect.StructField{
					Name: t.Field(i).Name,
					Type: t.Field(i).Type,
				})
			regularFields = append(regularFields,
				FieldInfo{
					Name: t.Field(i).Name,
					Type: t.Field(i).Type,
				})

		} else if v.Kind() == reflect.Struct {
			newStructFields = append(newStructFields,
				reflect.StructField{
					Name: t.Field(i).Name,
					Type: reflect.TypeOf(sql.NullString{}),
					Tag:  t.Field(i).Tag,
				})
			jsonFields = append(jsonFields,
				FieldInfo{
					Name: t.Field(i).Name,
					Type: t.Field(i).Type,
				})
		} else {
			panic("Error Creating Flat Struct Type: typ not valid. Can only accept sql type or nested struct (for json)")
		}
	}
	newStructType := reflect.StructOf(newStructFields)
	return newStructType, jsonFields, regularFields
}

// func QueryWithMultipleNamedParams(query string, params []sql.NamedArg, anonstruct interface{}) (interface{}, string, error) {
// 	// Convert Named Params to Interface so they can be passed to Query
// 	var paramsInterface []interface{}
// 	for _, param := range params {
// 		paramsInterface = append(paramsInterface, param)
// 	}
// 	// Execute Query
// 	rows, err := Sqlite.Query(query, paramsInterface...)
// 	if err != nil {
// 		fmt.Println(err)
// 		return nil, "Error Executing Query", err
// 	}
// 	// ~~
// 	flattenedStruct, jsonOutputs := flattenAndReturn(anonStruct)
// 	numFields := flattenedStruct.NumField()
//
// 	dataType := flattenedStruct
//
// 	// Create a slice to hold the pointers
// 	slicePointers := make([]interface{}, numFields)
//
// 	results := reflect.MakeSlice(reflect.TypeOf(slicePointers), 0, 0)
//
// 	for rows.Next() {
// 		// Create a new instance of the struct type
// 		newStructPtr := reflect.New(dataType).Elem()
//
// 		slicePointers := make([]interface{}, numFields)
//
// 		for i := 0; i < numFields; i++ {
// 			slicePointers[i] = newStructPtr.Field(i).Addr().Interface()
// 		}
//
// 		err := rows.Scan(slicePointers...)
// 		if err != nil {
// 			fmt.Println("error scanning rows into struct: ", err)
// 			return nil, "Error scanning rows into struct", err
// 		}
//
// 		for _, v := range jsonOutputs {
// 			jsonString := newStructPtr.FieldByName(v.Name).FieldByName("String").Interface()
//
// 			// Create a new instance of the slice type
// 			container := reflect.New(v.Type).Elem()
//
// 			// Since containerValue is already the correct type, assign it directly to concrete
// 			concrete := container.Addr().Interface()
//
// 			jsonAsserted, ok := jsonString.(string)
// 			if !ok {
// 				panic("not a valid json string")
// 			}
//
// 			////////////// Pretty Print JSON ////////////////////////
// 			// var jsonObj map[string]interface{}
// 			err := json.Unmarshal([]byte(jsonAsserted), concrete)
// 			if err != nil {
// 				log.Fatalf("Error unmarshalling JSON: %v", err)
// 			}
//
// 			// Marshal the JSON object with indentation
// 			prettyJSON, err := json.MarshalIndent(concrete, "", "  ")
// 			if err != nil {
// 				log.Fatalf("Error marshalling JSON: %v", err)
// 			}
//
// 			// Convert bytes.Buffer to string for logging
// 			var prettyString bytes.Buffer
// 			err = json.Indent(&prettyString, prettyJSON, "", "  ")
// 			if err != nil {
// 				log.Fatalf("Error indenting JSON: %v", err)
// 			}
//
// 			// Log the pretty-printed JSON string
// 			log.Println(prettyString.String())
//
// 			////////////// End Pretty Print JSON ////////////////////////
//
// 			dec := json.NewDecoder(strings.NewReader(jsonAsserted))
//
// 			if err := dec.Decode(concrete); err != nil && err != io.EOF {
// 				fmt.Println(err)
// 				panic("Error while decoding json")
// 			}
//
// 		}
// 		results = reflect.Append(results, newStructPtr)
// 	}
//
// 	sliceOfDataType := reflect.SliceOf(dataType)
// 	concreteDataSlice := reflect.MakeSlice(sliceOfDataType, 0, 0)
//
// 	// Check if the result is a slice
// 	// If it is, iterate through the slice and convert the items to the concrete type specified in routeConfig
// 	if results.Kind() == reflect.Slice {
// 		for i := 0; i < results.Len(); i++ {
// 			item := results.Index(i).Interface()
//
// 			// dereference the pointer to get the underlying struct for each slice item
// 			dereferencedItem := reflect.Indirect(reflect.ValueOf(item)).Interface()
//
// 			// Convert the dereferencedItem to the concrete type specified in routeConfig
// 			dereferencedItemValue := reflect.ValueOf(dereferencedItem)
//
// 			if dereferencedItemValue.Type().ConvertibleTo(dataType) {
// 				concrete := reflect.ValueOf(dereferencedItem).Convert(dataType)
// 				concreteDataSlice = reflect.Append(concreteDataSlice, concrete)
//
// 			} else {
// 				fmt.Println("Unexpected type")
// 			}
// 		}
// 	} else {
// 		fmt.Println("Unexpected result type")
// 	}
//
// 	return results.Interface(), "success", nil
// }

// func ExecTestWithNamedParams(query string, params []sql.NamedArg) (string, error) {
// 	var paramsInterface []interface{}
// 	for _, param := range params {
// 		paramsInterface = append(paramsInterface, param)
// 	}
//
// 	response, err := Sqlite.Exec(query, paramsInterface...)
// 	if err != nil {
// 		fmt.Println(err)
// 		return "Error Executing Exec Query", err
// 	}
//
// 	fmt.Println("response: ", response)
//
// 	return "Record created successfully", nil
// }

// Capitalize the first letter of a string
// func CapitalizeFirstLetter(s string) string {
// 	// Check if the string is empty
// 	if len(s) == 0 {
// 		return s
// 	}
//
// 	// Convert the string to a rune slice for proper handling of UTF-8 characters
// 	runes := []rune(s)
//
// 	// Capitalize the first rune
// 	runes[0] = unicode.ToUpper(runes[0])
//
// 	// Convert the rune slice back to a string
// 	return string(runes)
// }

// Helper function to check if a field is a SQL type
// func isSQLType(s reflect.Type) bool {
// 	sqlTypes := []reflect.Type{
// 		reflect.TypeOf(sql.Null[any]{}),
// 		reflect.TypeOf(sql.NullString{}),
// 		reflect.TypeOf(sql.NullByte{}),
// 		reflect.TypeOf(sql.NullInt64{}),
// 		reflect.TypeOf(sql.NullInt16{}),
// 		reflect.TypeOf(sql.NullBool{}),
// 		reflect.TypeOf(sql.NullTime{}),
// 		reflect.TypeOf(sql.NullInt32{}),
// 		reflect.TypeOf(sql.NullFloat64{}),
// 	}
//
// 	return slices.Contains(sqlTypes, s)
// }
//
// type FieldInfo struct {
// 	Name string
// 	Type reflect.Type
// }
//
////////////// Pretty Print JSON ////////////////////////
// var jsonObj map[string]interface{}
// err := json.Unmarshal([]byte(jsonAsserted), concrete)
// if err != nil {
// 	log.Fatalf("Error unmarshalling JSON: %v", err)
// }
//
// // Marshal the JSON object with indentation
// prettyJSON, err := json.MarshalIndent(concrete, "", "  ")
// if err != nil {
// 	log.Fatalf("Error marshalling JSON: %v", err)
// }
//
// // Convert bytes.Buffer to string for logging
// var prettyString bytes.Buffer
// err = json.Indent(&prettyString, prettyJSON, "", "  ")
// if err != nil {
// 	log.Fatalf("Error indenting JSON: %v", err)
// }
//
// // Log the pretty-printed JSON string
// log.Println(prettyString.String())

////////////// End Pretty Print JSON ////////////////////////

//
// func flattenAndReturn(s interface{}) (reflect.Type, []FieldInfo) {
// 	v := reflect.ValueOf(s)
// 	t := reflect.TypeOf(s)
// 	var newStructFields []reflect.StructField
//
// 	var jsonFields []FieldInfo
//
// 	for i := 0; i < t.NumField(); i++ {
// 		if isSQLType(t.Field(i).Type) {
// 			newStructFields = append(newStructFields,
// 				reflect.StructField{
// 					Name: t.Field(i).Name,
// 					Type: t.Field(i).Type,
// 				})
// 		} else if v.Kind() == reflect.Struct {
// 			newStructFields = append(newStructFields,
// 				reflect.StructField{
// 					Name: t.Field(i).Name,
// 					Type: reflect.TypeOf(sql.NullString{}),
// 					Tag:  t.Field(i).Tag,
// 				})
// 			jsonFields = append(jsonFields,
// 				FieldInfo{
// 					Name: t.Field(i).Name,
// 					Type: t.Field(i).Type,
// 				})
// 		} else {
// 			panic("Error: typ not valid. Can only accept sql type or nested struct (for json)")
// 		}
// 	}
// 	newStructType := reflect.StructOf(newStructFields)
// 	return newStructType, jsonFields
// }
