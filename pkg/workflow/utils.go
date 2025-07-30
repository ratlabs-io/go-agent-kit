package workflow

import (
	"fmt"
	"reflect"
)

// extractContent tries to extract text content from various data types
func extractContent(data interface{}) string {
	if data == nil {
		return ""
	}

	// Try to access Content field using reflection
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() == reflect.Struct {
		contentField := v.FieldByName("Content")
		if contentField.IsValid() && contentField.Kind() == reflect.String {
			return contentField.String()
		}
	}

	// Fallback to string representation
	return fmt.Sprintf("%v", data)
}
