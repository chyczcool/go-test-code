package main

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/tanpopoycz/go-test-code/module/bulitin"
	"github.com/tanpopoycz/go-test-code/module/config"
	"github.com/tanpopoycz/go-test-code/module/log"
)

func Test_log() {
	log.Test_log()
}

func Test_path() {
	// 定义一个文件路径
	filePath := "../xxx/ss/example."

	// 获取文件所在目录
	dir := filepath.Dir(filePath)
	fmt.Println("Directory:", dir)

	// 获取文件名
	filename := filepath.Base(filePath)
	fmt.Println("Filename:", filename)

	// 获取文件拓展名
	extension := filepath.Ext(filename)
	fmt.Println("Extension:", extension)
}

func Test_config() {
	config.Init("config", "json", "./module/config")
	fmt.Printf("log.level = %s\n", config.Test_viper_getstr("log.level"))
	fmt.Printf("log.file = %s\n", config.Test_viper_getstr("log.file"))
	fmt.Printf("log.backup_dir = %s\n", config.Test_viper_getstr("log.backup_dir"))

	if str := config.Test_viper_getstr("log.rotate_size"); str != "" {
		fmt.Printf("log.rotate_size = %s\n", str)
	}

	fmt.Printf("test.int = %d\n", config.Test_viper_getInt("test.int"))

}

func Test_exec() {
	cmd := exec.Command("./test")
	fmt.Printf("Running command and waiting for it to finish...\n")
	err := cmd.Run()
	fmt.Printf("Command finished with error: %v\n", err)
}

type T struct {
	x int
	y int
}

var dst_t = T{1, 2}

func Test_struct(t *T) {
	t = &dst_t
}

func GetLocalIP() ([]string, error) {
	var ipList []string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ipList, err
	}

	for _, address := range addrs {
		// 检查是否是IP地址
		var ip net.IP
		switch v := address.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}

		if ip == nil || ip.IsLoopback() {
			continue
		}

		ip = ip.To4()
		if ip == nil {
			continue // 不是IPv4地址
		}

		ipList = append(ipList, ip.String())

	}

	return ipList, nil
}

func Test_context() {
	// 创建一个空的背景 Context
	ctx, cancel := context.WithCancel(context.Background())

	// 使用该 Context 运行一个 goroutine
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				// 如果 Context 被取消，则退出 goroutine
				fmt.Println("Goroutine cancelled")
				return
			default:
				// 模拟一些工作
				time.Sleep(time.Second)
				fmt.Println("Goroutine working")
			}
		}
	}(ctx)

	// 等待一段时间
	time.Sleep(3 * time.Second)

	// 取消 Context
	fmt.Println("Cancelling the Context")

	cancel() // 手动取消 Context

	// 等待一段时间，观察 goroutine 是否退出
	time.Sleep(3 * time.Second)
}

// group集
type Group struct {
	Name      string   //group名称
	Consumers []string //group拥有的消费者列表
}

type Streams struct {
	Name   string  //streams name
	Groups []Group //
}

func NewStreams() *Streams {
	return new(Streams)
	//s.Groups = make([]Group, 0)
}

type MyInterface interface {
	fmt.Stringer
	MyString() string
}

type Tt struct {
	MyInterface

	I int
	J string
}

func (t Tt) String() string {
	return fmt.Sprintf("Tt{I:%d, J:%s}", t.I, t.J)
}

func (t Tt) MyString() string {
	return fmt.Sprintf("MyString: Tt{I:%d, J:%s}", t.I, t.J)
}

func main() {
	// t := Tt{I: 100, J: "abc"}

	// fmt.Printf("%t", true)
	// var i any = &t
	// v, ok := i.(MyInterface)
	// if ok {
	// 	fmt.Printf("t --> %s", v)
	// 	fmt.Print(v.MyString())
	// } else {
	// 	fmt.Printf("t --> null")
	// }

	//json.Test()

	bulitin.PrintType()

}
