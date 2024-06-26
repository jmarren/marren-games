package routers

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Utility function to convert query parameter from string to specified type
func ConvertType(value string, targetType reflect.Kind) (interface{}, error) {
	switch targetType {
	case reflect.Int:
		result, err := strconv.Atoi(value)
		if err != nil {
			fmt.Println("error converting string to int", err)
			fmt.Println("** Trying to convert ", value, " to int")
			return nil, err
		}
		return result, nil
	case reflect.String:
		return value, nil
	// Add more type cases as needed (e.g., float64, bool, etc.)
	default:
		return nil, fmt.Errorf("unsupported type: %s", targetType)
	}
}

// Function to convert a struct with sql.Null types to a struct with primitive types
func SimplifySqlResult(data interface{}) interface{} {
	// Get the type of the input struct
	dataType := reflect.TypeOf(data)
	dataValue := reflect.ValueOf(data)

	// Create a new struct type to hold the simplified fields
	simplifiedType := reflect.StructOf(GetSimplifiedFields(dataType))
	simplifiedInstance := reflect.New(simplifiedType).Elem()
	// Iterate over the fields of the original struct
	for i := 0; i < dataType.NumField(); i++ {
		field := dataType.Field(i)
		value := dataValue.Field(i)

		// Handle sql.NullString
		if field.Type == reflect.TypeOf(sql.NullString{}) {
			if value.FieldByName("Valid").Bool() {
				simplifiedInstance.Field(i).SetString(value.FieldByName("String").String())
			} else {
				simplifiedInstance.Field(i).SetString("")
			}
		}

		// Handle sql.NullInt64
		if field.Type == reflect.TypeOf(sql.NullInt64{}) {
			if value.FieldByName("Valid").Bool() {
				simplifiedInstance.Field(i).SetInt(value.FieldByName("Int64").Int())
			} else {
				simplifiedInstance.Field(i).SetInt(0)
			}
		}

		// Handle sql.NullFloat64
		if field.Type == reflect.TypeOf(sql.NullFloat64{}) {
			if value.FieldByName("Valid").Bool() {
				simplifiedInstance.Field(i).SetFloat(value.FieldByName("Float64").Float())
			} else {
				simplifiedInstance.Field(i).SetFloat(0.0)
			}
		}

		// Handle sql.NullBool
		if field.Type == reflect.TypeOf(sql.NullBool{}) {
			if value.FieldByName("Valid").Bool() {
				simplifiedInstance.Field(i).SetBool(value.FieldByName("Bool").Bool())
			} else {
				simplifiedInstance.Field(i).SetBool(false)
			}
		}
	}
	return simplifiedInstance.Interface()
}

// Helper function to get the simplified fields
func GetSimplifiedFields(dataType reflect.Type) []reflect.StructField {
	var fields []reflect.StructField

	for i := 0; i < dataType.NumField(); i++ {
		field := dataType.Field(i)

		var newField reflect.StructField
		switch field.Type {
		case reflect.TypeOf(sql.NullString{}):
			newField = reflect.StructField{
				Name: field.Name,
				Type: reflect.TypeOf(""),
				Tag:  field.Tag,
			}
		case reflect.TypeOf(sql.NullInt64{}):
			newField = reflect.StructField{
				Name: field.Name,
				Type: reflect.TypeOf(int64(0)),
				Tag:  field.Tag,
			}
		case reflect.TypeOf(sql.NullFloat64{}):
			newField = reflect.StructField{
				Name: field.Name,
				Type: reflect.TypeOf(float64(0)),
				Tag:  field.Tag,
			}
		case reflect.TypeOf(sql.NullBool{}):
			newField = reflect.StructField{
				Name: field.Name,
				Type: reflect.TypeOf(false),
				Tag:  field.Tag,
			}
		default:
			newField = reflect.StructField{
				Name: field.Name,
				Type: field.Type,
				Tag:  field.Tag,
			}
		}

		fields = append(fields, newField)
	}

	return fields
}

// Function to convert the simplified results to a string
func ResultsToString(simplifiedResults interface{}) string {
	v := reflect.ValueOf(simplifiedResults)
	if v.Kind() != reflect.Slice {
		return ""
	}

	var builder strings.Builder
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		builder.WriteString("{")
		for j := 0; j < elem.NumField(); j++ {
			field := elem.Field(j)
			builder.WriteString(fmt.Sprintf("%s: %v", elem.Type().Field(j).Name, field.Interface()))
			if j < elem.NumField()-1 {
				builder.WriteString(", ")
			}
		}
		builder.WriteString("}\n")
	}
	return builder.String()
}

func PrintStruct(s interface{}) {
	val := reflect.ValueOf(s)
	typ := reflect.TypeOf(s)

	if val.Kind() == reflect.Struct {
		fmt.Printf("Struct type: %s\n", typ)
		for i := 0; i < val.NumField(); i++ {
			fieldName := typ.Field(i).Name
			fieldValue := val.Field(i).Interface()
			fmt.Printf("%s: %v\n", fieldName, fieldValue)
		}
	} else {
		fmt.Println("Provided value is not a struct")
	}
}