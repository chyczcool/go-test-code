// reflect reflect测试，实现结构数据的互转、序列化和反序列化
package reflect

import (
	"errors"
	"fmt"
	"reflect"
)

/* BEGIN:实现一个接口方法，可以实现结构体和map之间的互转 */

// StructToMap 将结构体转换为map[string]interface{}， 值只支持基础类型
func StructToMap(obj interface{}) (map[string]interface{}, error) {
	//obj可能是指针
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// 确保obj是一个结构体
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected a struct, got %s", val.Kind())
	}

	out := make(map[string]interface{})
	for i := 0; i < val.NumField(); i++ {
		fieldType := val.Type().Field(i) //字段类型
		fieldValue := val.Field(i)       //字段值

		//跳过不可导出的elem
		if !fieldType.IsExported() {
			fmt.Printf("WARN field[%s.%s] is not exported.\n", val.Type().Name(), fieldType.Name)
			continue
		}

		tag := fieldType.Tag.Get("json") // 尝试获取json标签
		if tag == "" {
			tag = fieldType.Name
		}

		out[tag] = fieldValue.Interface()
	}
	return out, nil
}

// MapToStruct 将map[string]interface{}转换为结构体
func MapToStruct(m map[string]interface{}, obj interface{}) error {
	//obj参数只能是指针
	val := reflect.ValueOf(obj)
	if val.Kind() != reflect.Ptr {
		return errors.New("obj param must be pointer")
	}

	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("expected a pointer to struct, got %s", val.Kind())
	}

	for i := 0; i < val.NumField(); i++ {
		fieldType := val.Type().Field(i) //字段类型
		fieldValue := val.Field(i)       //字段值

		//跳过不可导出的elem
		if !fieldType.IsExported() {
			fmt.Printf("WARN field[%s.%s] is not exported.\n", val.Type().Name(), fieldType.Name)
			continue
		}

		// 查找map中的key，key可能是字段名或json标签
		var key string
		if tag := fieldType.Tag.Get("json"); tag != "" {
			key = tag
		} else {
			key = fieldType.Name
		}

		mapValue, ok := m[key]
		if !ok {
			fmt.Printf("WARN struct filed[%s.%s] not found in map.\n", val.Type().Name(), key)
			continue // 如果map中没有这个key，则跳过
		}

		// 设置结构体字段的值
		fieldValue.Set(reflect.ValueOf(mapValue))
	}
	return nil
}

/* END:实现一个接口方法，可以实现结构体和map之间的互转 */

type ttt struct {
	X int
	Y int
}

func Test() {
	t := reflect.TypeFor[int]()
	fmt.Println(t)

	var i int = 123
	if reflect.TypeOf(i).Kind() == reflect.Int {
		fmt.Println(reflect.TypeOf(i))
	}
	var po = ttt{X: 1, Y: 1}
	name := reflect.TypeOf(po).Name()
	fmt.Println("struct name: ", name)

	name1 := reflect.ValueOf(po).Type().Name()
	fmt.Println("struct name: ", name1)
}
