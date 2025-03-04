// Package settings provides ...
package settings

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// Init 初始化配置
func Init() (err error) {
	viper.SetConfigName("config") // 配置文件名称（不带后缀）
	viper.SetConfigType("yaml")   // 指定配置文件类型
	viper.AddConfigPath("./conf") // 指定查找配置文件的路径（这里使用相对路径）
	err = viper.ReadInConfig()    // 读取配置信息
	if err != nil {
		// 读取配置信息失败
		fmt.Printf("viper.ReadInConfig() failed, err:%v\n", err)
		return
	}
	viper.WatchConfig() // 热加载配置
	// 热加载触发时的回调函数
	viper.OnConfigChange(func(in fsnotify.Event) {
		fmt.Println("配置文件更新了...")
	})
	return
}
