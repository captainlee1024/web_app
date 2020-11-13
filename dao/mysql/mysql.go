// Package mysql provides ...
package mysql

import (
	"fmt"

	// 匿名应用驱动，会使用里面的Init()函数
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var db *sqlx.DB

// Init 初始化 MySQL 连接
func Init() (err error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True",
		viper.GetString("mysql.user"),
		viper.GetString("mysql.password"),
		viper.GetString("mysql.host"),
		viper.GetInt("mysql.port"),
		viper.GetString("mysql.dbname"),
	)

	// 也可以使用 MustConnect() ，连接不成功就 panic
	db, err = sqlx.Connect("mysql", dsn)
	if err != nil {
		zap.L().Error("connect DB failed", zap.Error(err))
		return
	}

	db.SetMaxOpenConns(viper.GetInt("mysql.max_open_conns")) // 设置最大连接数
	db.SetMaxIdleConns(viper.GetInt("mysql.max_idle_conns")) // 设置最大空闲连接数

	return
}

// Close 关闭 MySQL 连接
// 我们声明的全局变量是小写的，所以需要写一个外面可以调的函数去关闭
func Close() {
	_ = db.Close()
}
