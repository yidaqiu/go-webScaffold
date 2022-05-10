package logger

import (
	"ginframe/webScaffold/settings"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

//var Logger *zap.Logger
//var SuggarLogger *zap.SugaredLogger

func Init(loggerConfig *settings.LogConfig) (err error) {
	err = InitLogger(loggerConfig)
	defer zap.L().Sync()
	return
}

func InitLogger(loggerConfig *settings.LogConfig) (err error) {
	writeSyncer := getLogWriter(loggerConfig)
	encoder := getEncoder()

	var l = new(zapcore.Level)
	if err = l.UnmarshalText([]byte(loggerConfig.Level)); err != nil {
		return
	}
	core := zapcore.NewCore(encoder, writeSyncer, l)

	logger := zap.New(core)
	// 日志加上调用方法
	// logger := zap.New(core, zap.AddCaller())
	//Logger = logger
	//SuggarLogger = logger.Sugar()
	zap.ReplaceGlobals(logger)

	return
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	// 修改时间编码器，设置时间格式
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	// 在日志文件中使用大写字母记录日志级别
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewJSONEncoder(encoderConfig)
}

func getLogWriter(loggerConfig *settings.LogConfig) zapcore.WriteSyncer {
	lumberjackLogger := &lumberjack.Logger{
		Filename:   loggerConfig.Filename,   // 日志文件的位置
		MaxSize:    loggerConfig.MaxSize,    // 备份文件最大多大（以MB为单位）
		MaxBackups: loggerConfig.MaxBackups, // 最大备份数量
		MaxAge:     loggerConfig.MaxAge,     // 最大备份天数
		Compress:   false,                   // 是否压缩
	}
	return zapcore.AddSync(lumberjackLogger)
}

// gin 的 logger 中间件
func GinLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next()

		cost := time.Since(start)
		// 结构化的日志，明确类型，减少内存分配。
		logger.Info(
			path,
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()),
			zap.Duration("cost", cost),
		)
		// loger.Sugar.Info()  参数是interface，所以使用方便，但是性能下降。
	}
}

func GinRecovery(logger *zap.Logger, stack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
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
					logger.Error(c.Request.URL.Path,
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)

					c.Error(err.(error))
					c.Abort()
					return
				}

				if stack {
					logger.Error("[Recovery from panic]",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
						zap.String("stack", string(debug.Stack())),
					)
				} else {
					logger.Error("[Recovery from panic]",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)))
				}

				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()

		c.Next()
	}
}
