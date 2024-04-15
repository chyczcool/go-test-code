// test http client server transport
package http

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

var MyClient = http.Client{
	Timeout: time.Second * 5,
}

func GetBaidu() {
	rsp, err := MyClient.Get("http://baidu.com")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rsp.Body.Close()

	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	for k, v := range rsp.Header {
		fmt.Println(k, ": ", v)
	}
	fmt.Println("rsp body ->\n", string(body))
}
