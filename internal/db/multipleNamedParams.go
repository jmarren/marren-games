package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"slices"
	"strings"
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

	flattenedStruct, jsonOutputs := flattenAndReturn(anonStruct)
	numFields := flattenedStruct.NumField()
	fmt.Println("flattenedStruct.NumField(): ", numFields)

	dataType := flattenedStruct
	fmt.Println("flattenedStruct: ", reflect.TypeOf(flattenedStruct))
	fmt.Println("jsonOutputs: ", jsonOutputs)
	// Create a slice to hold the pointers
	slicePointers := make([]interface{}, numFields)

	results := reflect.MakeSlice(reflect.TypeOf(slicePointers), 0, 0)

	for rows.Next() {
		// Create a new instance of the struct type
		newStructPtr := reflect.New(dataType).Elem()
		fmt.Println("newStructPtr: ", newStructPtr)

		fmt.Println(newStructPtr.Type())

		slicePointers := make([]interface{}, numFields)

		for i := 0; i < numFields; i++ {
			slicePointers[i] = newStructPtr.Field(i).Addr().Interface()
		}

		fmt.Println("Starting Scan")

		err := rows.Scan(slicePointers...)
		if err != nil {
			fmt.Println("error scanning rows into struct: ", err)
			return nil, "Error scanning rows into struct", err
		}

		fmt.Println("newStructPtr[i] : ", newStructPtr)

		for _, v := range jsonOutputs {
			jsonString := newStructPtr.FieldByName(v.Name).FieldByName("String").Interface()
			fmt.Print("\n\n------------------")
			fmt.Println("type of jsonString: ", reflect.TypeOf(jsonString))
			fmt.Print("\n\n------------------")
			fmt.Println("--------- jsonString: ", reflect.ValueOf(jsonString))
			fmt.Println("------------ v.Type: ", v.Type)
			fmt.Print("\n\n")

			// Create a new instance of the slice type
			container := reflect.New(v.Type).Elem()
			// containerValue := reflect.New(reflect.StructOf(v.Type)).Elem()
			// containerSlice := reflect.MakeSlice(reflect.SliceOf(v.Type), 0, 0)
			// ptr := reflect.PointerTo(containerSlice.Type())

			// fmt.Println("type of containerSlice: ", reflect.TypeOf(containerSlice))
			// fmt.Println("containerValue (initial): ", containerValue)

			// Since containerValue is already the correct type, assign it directly to concrete
			concrete := container.Addr().Interface()
			fmt.Println("concrete (initial): ", concrete)

			jsonAsserted, ok := jsonString.(string)
			if !ok {
				panic("not a valid json string")
			}

			dec := json.NewDecoder(strings.NewReader(jsonAsserted))

			if err := dec.Decode(concrete); err != nil && err != io.EOF {
				fmt.Println(err)
				panic("Error while decoding json")
			}

			fmt.Println("container: ", container)

		}
		results = reflect.Append(results, newStructPtr)
	}

	sliceOfDataType := reflect.SliceOf(dataType)
	concreteDataSlice := reflect.MakeSlice(sliceOfDataType, 0, 0)

	// Check if the result is a slice
	// If it is, iterate through the slice and convert the items to the concrete type specified in routeConfig
	if results.Kind() == reflect.Slice {
		for i := 0; i < results.Len(); i++ {
			item := results.Index(i).Interface()

			// dereference the pointer to get the underlying struct for each slice item
			dereferencedItem := reflect.Indirect(reflect.ValueOf(item)).Interface()

			// Convert the dereferencedItem to the concrete type specified in routeConfig
			dereferencedItemValue := reflect.ValueOf(dereferencedItem)

			if dereferencedItemValue.Type().ConvertibleTo(dataType) {
				concrete := reflect.ValueOf(dereferencedItem).Convert(dataType)
				concreteDataSlice = reflect.Append(concreteDataSlice, concrete)

			} else {
				fmt.Println("Unexpected type")
			}
		}
	} else {
		fmt.Println("Unexpected result type")
	}

	fmt.Println("concreteDataSlice", concreteDataSlice)

	return results.Interface(), "success", nil
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

func GetNumFields(s reflect.Value) int {
	j := s.NumField()

	sqlTypes := []reflect.Type{
		reflect.TypeOf(sql.Null[any]{}),
		reflect.TypeOf(sql.NullString{}),
		reflect.TypeOf(sql.NullByte{}),
		reflect.TypeOf(sql.NullInt64{}),
		reflect.TypeOf(sql.NullInt16{}),
		reflect.TypeOf(sql.NullBool{}),
		reflect.TypeOf(sql.NullTime{}),
		reflect.TypeOf(sql.NullInt32{}),
		reflect.TypeOf(sql.NullFloat64{}),
	}

	for i := 0; i < j; i++ {
		current := s.Field(i)
		if !slices.Contains(sqlTypes, current.Type()) {
			return j + GetNumFields(current)
		}
	}
	return j
}

func FillPointersSlice(newStructPtr reflect.Value) []interface{} {
	// j := s.NumField()
	// slicePointers := make([]interface{}, numFieldsWithNested)

	slicePointers := []interface{}{}

	sqlTypes := []reflect.Type{
		reflect.TypeOf(sql.Null[any]{}),
		reflect.TypeOf(sql.NullString{}),
		reflect.TypeOf(sql.NullByte{}),
		reflect.TypeOf(sql.NullInt64{}),
		reflect.TypeOf(sql.NullInt16{}),
		reflect.TypeOf(sql.NullBool{}),
		reflect.TypeOf(sql.NullTime{}),
		reflect.TypeOf(sql.NullInt32{}),
		reflect.TypeOf(sql.NullFloat64{}),
	}

	numFields := newStructPtr.NumField()
	// numNestedFields := GetNumFields(newStructPtr)

	for i := 0; i < numFields; i++ {
		fmt.Println("i:", i)
		if slices.Contains(sqlTypes, newStructPtr.Field(i).Type()) {
			fmt.Println("is sql type...")
			slicePointers = append(slicePointers, newStructPtr.Field(i).Addr().Interface())
		} else {
			fmt.Println("is not sql type...")
			slicePointers = append(slicePointers, FillPointersSlice(newStructPtr.Field(i))...)
		}
	}

	return slicePointers
}

// Helper function to check if a field is a SQL type
func isSQLType(s reflect.Type) bool {
	sqlTypes := []reflect.Type{
		reflect.TypeOf(sql.Null[any]{}),
		reflect.TypeOf(sql.NullString{}),
		reflect.TypeOf(sql.NullByte{}),
		reflect.TypeOf(sql.NullInt64{}),
		reflect.TypeOf(sql.NullInt16{}),
		reflect.TypeOf(sql.NullBool{}),
		reflect.TypeOf(sql.NullTime{}),
		reflect.TypeOf(sql.NullInt32{}),
		reflect.TypeOf(sql.NullFloat64{}),
	}

	if !slices.Contains(sqlTypes, s) {
		return false
	}
	return true
}

type FieldInfo struct {
	Name string
	Type reflect.Type
}

func flattenAndReturn(s interface{}) (reflect.Type, []FieldInfo) {
	v := reflect.ValueOf(s)
	t := reflect.TypeOf(s)
	// newStructFields := make([]reflect.StructField, t.NumField())
	var newStructFields []reflect.StructField

	var jsonFields []FieldInfo

	// for i := 0; i < t.NumField(); i++ {
	// fmt.Println("t.Field(i).Name: ", t.Field(i).Name)
	// fmt.Println("reflect.typeof(t.Field(i)): ", reflect.TypeOf(t.Field(i)))
	// fmt.Println("isSQLType: ", isSQLType(t.Field(i).Type))
	//
	for i := 0; i < t.NumField(); i++ {
		if isSQLType(t.Field(i).Type) {
			newStructFields = append(newStructFields,
				reflect.StructField{
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
			panic("Error: typ not valid. Can only accept sql type or nested struct (for json)")
		}
	}
	// }
	fmt.Print("\n\n\n------------------")
	fmt.Println("newStructFields: ", newStructFields)
	fmt.Print("\n\n\n------------------")
	newStructType := reflect.StructOf(newStructFields)
	return newStructType, jsonFields
}

// func dereferenceResult(result interface{})
