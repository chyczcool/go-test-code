package testjson

import "encoding/json"

type TestJsonStrcut struct {
	FieldA string
	FieldB int
}

func (t *TestJsonStrcut) JsonString() ([]byte, error) {
	return json.Marshal(*t)
}
