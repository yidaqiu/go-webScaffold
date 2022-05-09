package main

import (
	"context"
	"fmt"
	"ginframe/webScaffold/dao/mysql"
	"ginframe/webScaffold/dao/redis"
	"ginframe/webScaffold/logger"
	"ginframe/webScaffold/routes"
	"ginframe/webScaffold/settings"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/viper"

	"go.uber.org/zap"
)

func main() {
	// 加载配置
	if err := settings.Init(); err != nil {
		fmt.Printf("init settings failed, err:%v\n", err)
		return
	}
	// 初始化日志
	if err := logger.Init(); err != nil {
		fmt.Printf("init settings failed, err:%v\n", err)
		return
	}
	//defer zap.L().Sync()
	zap.L().Debug("log init success")
	// 初始化mysql 连接
	if err := mysql.Init(); err != nil {
		fmt.Printf("init settings failed, err:%v\n", err)
		return
	}
	//defer mysql.Close()
	// 初始化redis 连接
	if err := redis.Init(); err != nil {
		fmt.Printf("init settings failed, err:%v\n", err)
		return
	}
	//redis.Close()
	// 注册路由
	r := routes.Setup()

	// 启动服务（优雅关机）
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", viper.GetInt("app.port")),
		Handler: r,
	}

	go func() {
		// 开启一个goroutine启动服务, 不然程序一直处理请求，不会往下走
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.S().Error("listen: %s\n", err)
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
	zap.S().Info("Shutdown Server ...")
	// 创建一个5秒超时的context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// 5秒内优雅关闭服务（将未处理完的请求处理完再关闭服务），超过5秒就超时退出
	if err := srv.Shutdown(ctx); err != nil {
		zap.S().Fatal("Server Shutdown: ", err)
	}

	zap.S().Info("Server exiting")
}