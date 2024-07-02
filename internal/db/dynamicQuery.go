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

	fmt.Println("rows: ", rows)

	// Craete a flattened version of anonStruct
	flattenedStructType, jsonOutputs, regularOutputs := CreateFlatStructType(anonStruct)
	numFields := flattenedStructType.NumField()

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

	}
	fmt.Printf("\njsonOutputs: %v\nnewStructPtr: %v", jsonOutputs, newStructPtr)
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
