package routers

import (
	"database/sql"
	"fmt"
	"reflect"
	"slices"
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

func SimplifySqlResult(structValue reflect.Value, simplifiedFields []reflect.StructField) interface{} {
	// Get the type of the input struct
	dataType := structValue.Type()

	// Create a new struct type to hold the simplified fields
	simplifiedType := reflect.StructOf(simplifiedFields)

	simplifiedInstance := reflect.New(simplifiedType).Elem()

	// Iterate over the fields of the original struct
	for i := 0; i < dataType.NumField(); i++ {
		field := dataType.Field(i)
		value := structValue.Field(i)

		if isSQLType(field.Type) {
			if value.FieldByName("Valid").Bool() {
				fieldName := reflect.Type(field.Type).Field(0).Name
				settingTo := value.FieldByName(fieldName)
				simplifiedInstance.Field(i).Set(settingTo)
			} else {
				simplifiedInstance.Field(i).SetString("")
			}
		} else {
			simplifiedInstance.Field(i).Set(structValue.FieldByName(field.Name))
		}
	}
	return simplifiedInstance.Interface()
}

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

	return slices.Contains(sqlTypes, s)
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
