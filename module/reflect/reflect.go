// reflect reflect测试，实现结构数据的互转、序列化和反序列化
package reflect

import (
	"fmt"
	"reflect"
)

/* BEGIN:实现一个接口方法，可以实现结构体和map之间的互转 */

// StructToMap 将结构体转换为map[string]interface{}
func StructToMap(obj interface{}) (map[string]interface{}, error) {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// 确保obj是一个结构体
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected a struct, got %s", val.Kind())
	}

	out := make(map[string]interface{})
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		tag := typ.Field(i).Tag.Get("json") // 假设我们使用json标签
		if tag == "" {
			tag = typ.Field(i).Name
		}
		out[tag] = field.Interface()
	}
	return out, nil
}

// MapToStruct 将map[string]interface{}转换为结构体
func MapToStruct(m map[string]interface{}, obj interface{}) error {
	val := reflect.ValueOf(obj).Elem()
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("expected a pointer to struct, got %s", val.Kind())
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		fieldType := typ.Field(i)
		fieldValue := val.Field(i)

		// 查找map中的key，key可能是字段名或json标签
		var key string
		if tag := fieldType.Tag.Get("json"); tag != "" {
			key = tag
		} else {
			key = fieldType.Name
		}

		mapValue, ok := m[key]
		if !ok {
			continue // 如果map中没有这个key，则跳过
		}

		// 设置结构体字段的值
		fieldValue.Set(reflect.ValueOf(mapValue))
	}
	return nil
}

/* END:实现一个接口方法，可以实现结构体和map之间的互转 */
