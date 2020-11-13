// Package logger provides ...
package logger

import (
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/natefinch/lumberjack"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 声明全局 logger
// 不推荐使用方式，因为在调用的时候需要 logger.logger 不是很方便
// 这里推荐使用另一种方式
//var logger *zap.Logger

// Init 初始化 Logger
// 定制 logger 并使用配置生成 logger　替换掉 zap 库里定义的全局 logger
func Init() (err error) {
	writerSyncer := getLogWriter(
		viper.GetString("log.filename"),
		viper.GetInt("log.max_size"),
		viper.GetInt("log.max_backups"),
		viper.GetInt("log.max_age"),
	)
	encoder := getEncoder()

	// 把 yaml 配置文件中 string　类型的 level 配置解析成 zap　日志库里面的 level 类型
	var l = new(zapcore.Level)
	err = l.UnmarshalText([]byte(viper.GetString("log.level")))
	if err != nil {
		return
	}
	core := zapcore.NewCore(encoder, writerSyncer, zapcore.DebugLevel)

	// 生成配置好的 logger
	lg := zap.New(core, zap.AddCaller())
	// 替换 zap 中的全局 logger ，这样在其他地方调用 zap.L() 使用 zap 里面定义的全局 logger 就是我们配置之后的 logger 了
	zap.ReplaceGlobals(lg)
	return
}

func getEncoder() zapcore.Encoder {
	// 配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:       "ts",
		LevelKey:      "level",
		NameKey:       "logger", // 名字是什么
		CallerKey:     "caller", // 调用者的名字
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.LowercaseLevelEncoder,
		//EncodeTime: zapcore.EpochTimeEncoder, // 默认的时间编码器
		EncodeTime:     zapcore.ISO8601TimeEncoder, // 修改之后的时间编码器
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 配置 JSON 编码器
	return zapcore.NewJSONEncoder(encoderConfig)

	// 配置 Console 编码器
	//return zapcore.NewConsoleEncoder(encoderConfig)
}

// 配置使用第三方日志分割归档 Lumberjack
func getLogWriter(filename string, maxSize, maxBackup, maxAge int) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   filename,  // 文件名称
		MaxSize:    maxSize,   // 单个文件最大 10 单位M
		MaxBackups: maxBackup, // 最大备份数量
		MaxAge:     maxAge,    // 最大备份天数（最多保存多少天的备份）
		//Compress:   false,        // 是否压缩（日志量比较大，或者磁盘空间有限可以开启，默认关闭）
	}
	return zapcore.AddSync(lumberJackLogger)
}

// GinLogger 接收gin框架默认的日志
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next()

		cost := time.Since(start)
		zap.L().Info(path,
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()),
			zap.Duration("cost", cost),
		)
	}
}

// GinRecovery recover掉项目可能出现的panic
func GinRecovery(stack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				if brokenPipe {
					zap.L().Error(c.Request.URL.Path,
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
					// If the connection is dead, we can't write a status to it.
					c.Error(err.(error)) // nolint: errcheck
					c.Abort()
					return
				}

				if stack {
					zap.L().Error("[Recovery from panic]",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
						zap.String("stack", string(debug.Stack())),
					)
				} else {
					zap.L().Error("[Recovery from panic]",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
				}
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}
