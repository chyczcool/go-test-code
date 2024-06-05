package bulitin

import "fmt"

func PrintType() {
	type myStruct struct {
		filedi int
		fileds string
	}

	v1 := 1
	v2 := "string"
	v3 := myStruct{filedi: v1, fileds: v2}

	fmt.Printf("v1:%T, v2:%T, v3:%T\n", v1, v2, v3)
}
