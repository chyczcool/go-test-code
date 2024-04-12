// package config 这是一个测试读取各种配置文件的测试模块
package config

import (
	"fmt"

	"github.com/spf13/viper"
)

func Test_viper() {
	// 设置配置文件路径
	viper.SetConfigName("config")   // 配置文件名（无扩展名）
	viper.SetConfigType("yaml")     // 配置文件类型
	viper.AddConfigPath("./config") // 配置文件所在路径

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// 配置文件未找到，但可能不是错误
			fmt.Println(err)
		} else {
			// 配置文件找到，但解析错误
			fmt.Println(err)
		}
	}

	// 获取配置项的值
	database := viper.GetString("database.user")
	fmt.Println("Database User:", database)
}
