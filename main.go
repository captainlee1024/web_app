// Package main provides ...
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/captainlee1024/web_app/dao/redis"
	"github.com/captainlee1024/web_app/logger"
	"github.com/captainlee1024/web_app/routes"
	"github.com/captainlee1024/web_app/settings"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// Go Web 开发通用的脚手架模板

func main() {
	// 1. 加载配置 一般会创建一个 settings 模块，使用 viper 去进行管理
	if err := settings.Init(); err != nil {
		fmt.Printf("init settings failed, err:%v\n", err)
		return
	}

	// 2. 初始化日志
	if err := logger.Init(); err != nil {
		// zap 初始化失败，这里不能使用 zap.L() 进行记录
		fmt.Printf("init logger failed, err:%v\n", err)
		return
	}
	// 延迟调用 刷新日志到文件
	defer zap.L().Sync()
	// 程序能走到这里说明 zap 日志配置已经成功，下面的都可以使用 zap.L()
	zap.L().Debug("logger init success...")

	/*
		// 3. 初始化 MySQL 连接
		if err := mysql.Init(); err != nil {
			fmt.Printf("init mysql failed, errL%v\n", err)
			return
		}
		defer mysql.Close()
	*/
	// 4. 初始化 Redis 连接
	if err := redis.Init(); err != nil {
		fmt.Printf("init redis failed, err:%v\n", err)
		return
	}
	defer redis.Close()

	// 5. 注册路由
	r := routes.Setup()

	// 6. 启动服务（优雅关机和平滑重启）
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", viper.GetInt("app.port")),
		Handler: r,
	}

	go func() {
		// 开启一个goroutine启动服务
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.L().Fatal("listen: ", zap.Error(err))
		}
	}()

	// 等待中断信号来优雅地关闭服务器，为关闭服务器操作设置一个5秒的超时
	quit := make(chan os.Signal, 1) // 创建一个接收信号的通道
	// kill 默认会发送 syscall.SIGTERM 信号
	// kill -2 发送 syscall.SIGINT 信号，我们常用的Ctrl+C就是触发系统SIGINT信号
	// kill -9 发送 syscall.SIGKILL 信号，但是不能被捕获，所以不需要添加它
	// signal.Notify把收到的 syscall.SIGINT或syscall.SIGTERM 信号转发给quit
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // 此处不会阻塞
	<-quit                                               // 阻塞在此，当接收到上述两种信号时才会往下执行
	zap.L().Info("Shutdown Server ...")
	// 创建一个5秒超时的context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// 5秒内优雅关闭服务（将未处理完的请求处理完再关闭服务），超过5秒就超时退出
	if err := srv.Shutdown(ctx); err != nil {
		zap.L().Fatal("Server Shutdown: ", zap.Error(err))
	}

	zap.L().Info("Server exiting")
}
