package reflection

import (
	"reflect"
	"unsafe"
)

func GetField(target any, fieldName string) reflect.Value {
	reflectionTarget := reflect.ValueOf(target)

	return reflect.Indirect(reflectionTarget).FieldByName(fieldName)
}

func SetField(target any, fieldName string, newValue any) {
	field := GetField(target, fieldName)

	fieldAddress := unsafe.Pointer(field.UnsafeAddr())
	settableField := reflect.NewAt(field.Type(), fieldAddress).Elem()

	newReflectionValue := reflect.ValueOf(newValue)
	settableField.Set(newReflectionValue)
}
