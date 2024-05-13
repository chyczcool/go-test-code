package json

import (
	"encoding/json"
	"fmt"
)

type TestJsonStruct struct {
	fmt.Stringer `json:"-"`

	Id   uint
	Name string
}

// Struct2JsonString 结构体转json字符串
func Struct2JsonString(obj interface{}) (string, error) {
	if b, err := json.Marshal(obj); err != nil {
		return "", err
	} else {
		return string(b), nil
	}
}

// JsonString2Struct json转结构体
func JsonString2Struct(jsonStr string, obj interface{}) error {
	return json.Unmarshal([]byte(jsonStr), obj)
}

func (st TestJsonStruct) String() string {
	return fmt.Sprintf("Id: %d, Name: %s", st.Id, st.Name)
}

func Test() {
	var jstr = `{"Id":100,"Nam":"xiaoming", "test": 123123}`

	var st TestJsonStruct

	if err := JsonString2Struct(jstr, &st); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("st -> %s\n", st)

}
