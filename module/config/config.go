// package config 这是一个测试读取各种配置文件的测试模块
package config

import (
	"fmt"

	"github.com/spf13/viper"
)

var myConfig = viper.New()

func Init(fileName string, fileType string, filePath string) {

	// 设置配置文件路径
	myConfig.SetConfigName(fileName) // 配置文件名（无扩展名）
	myConfig.SetConfigType(fileType) // 配置文件类型
	myConfig.AddConfigPath(filePath) // 配置文件所在路径

	//设置默认值
	myConfig.SetDefault("log.level", "info")
	myConfig.SetDefault("log.file", "./test.log")
	myConfig.SetDefault("log.backup_dir", "./backup")
	myConfig.SetDefault("log.rotate_size", "10MB")
	myConfig.SetDefault("go2rtc.bin", "go2rtc")
	myConfig.SetDefault("go2rtc.config", "./module/config/go2rtc.yaml")

	// 读取配置文件
	if err := myConfig.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			//不存在配置文件使用默认值就行

		} else {
			// 配置文件找到，但解析错误
			fmt.Println(err)
		}
	}
}

func Test_viper_getstr(key string) string {

	// 获取配置项的值
	return myConfig.GetString(key)

}

func Test_viper_getInt(key string) int {

	// 获取配置项的值
	return myConfig.GetInt(key)

}
