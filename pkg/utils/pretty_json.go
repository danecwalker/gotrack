package utils

import (
	"fmt"
	"reflect"
	"sort"
)

type Colors string

const (
	Black   Colors = "\033[30m"
	Red     Colors = "\033[31m"
	Green   Colors = "\033[32m"
	Yellow  Colors = "\033[33m"
	Blue    Colors = "\033[34m"
	Magenta Colors = "\033[35m"
	Cyan    Colors = "\033[36m"
	White   Colors = "\033[37m"
	Muted   Colors = "\033[90m"
	Reset   Colors = "\033[0m"
)

func Color(color Colors, v string) string {
	return fmt.Sprintf("%s%s%s", color, v, Reset)
}
func ColorF(color Colors, formatString string, v any) string {
	return fmt.Sprintf("%s%s%s", color, fmt.Sprintf(formatString, v), Reset)
}

func PrettyJson(value interface{}) string {
	switch v := value.(type) {
	case nil:
		return Color(Magenta, "null")
	case bool:
		if v {
			return Color(Yellow, "true")
		}
		return Color(Yellow, "false")
	case int:
		return ColorF(Cyan, "%d", v)
	case float64:
		return ColorF(Cyan, "%g", v)
	case string:
		return Color(Green, encodeString(v))
	case []interface{}:
		return encodeArray(v)
	case map[string]interface{}:
		return encodeObject(v)
	default:
		// Handle other types, e.g., structs
		val := reflect.ValueOf(value)
		switch val.Kind() {
		case reflect.Struct:
			return encodeObjectFromStruct(val)
		case reflect.Pointer:
			switch val.Elem().Kind() {
			case reflect.Struct:
				return encodeObjectFromStruct(val.Elem())
			default:
				return Color(Magenta, "null")
			}
		case reflect.String:
			return Color(Green, encodeString(val.String()))
		default:
			return Color(Magenta, "null")
		}
	}
}

// encodeString encodes a string value.
func encodeString(s string) string {
	return `"` + s + `"`
}

// encodeArray encodes an array.
func encodeArray(arr []interface{}) string {
	result := "["
	for i, v := range arr {
		if i > 0 {
			result += ", "
		}
		result += PrettyJson(v)
	}
	result += "]"
	return result
}

// encodeObject encodes a map[string]interface{}.
func encodeObject(obj map[string]interface{}) string {
	// Sort keys alphabetically
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	result := Color(Muted, "{")
	first := true
	for _, k := range keys {
		v := obj[k]
		if !first {
			result += ", "
		}
		first = false
		result += encodeString(k) + ": " + PrettyJson(v)
	}
	result += Color(Muted, "}")
	return result
}

// encodeObjectFromStruct encodes a struct.
func encodeObjectFromStruct(val reflect.Value) string {
	result := Color(Muted, "{")
	typ := val.Type()
	first := true

	// Collect and sort field names
	keys := make([]string, 0, val.NumField())
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		tag := field.Tag.Get("json")
		if tag == "-" {
			// Skip unexported or ignored fields
			continue
		} else if tag == "" {
			keys = append(keys, field.Name)
			continue
		}
		keys = append(keys, tag)
	}
	sort.Strings(keys)

	for _, k := range keys {
		for i := 0; i < val.NumField(); i++ {
			field := typ.Field(i)
			tag := field.Tag.Get("json")
			if tag == k {
				if !first {
					result += ", "
				}
				first = false
				result += encodeString(tag) + ": " + PrettyJson(val.Field(i).Interface())
				break
			} else if field.Name == k {
				if !first {
					result += ", "
				}
				first = false
				result += encodeString(field.Name) + ": " + PrettyJson(val.Field(i).Interface())
				break
			}
		}
	}
	result += Color(Muted, "}")
	return result
}
