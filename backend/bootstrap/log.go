package bootstrap

import (
	"os"
	"path/filepath"

	"github.com/pudongping/wx-graph-crawl/backend/configs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func InitZapLog(cfg *configs.Config) *zap.Logger {
	// 编码器，如何写入日志
	encoder := getEncoder(cfg.App.Mode)
	// 将日志写到哪里去
	writeSyncer := getLogWriter(cfg)
	// 日志级别
	logLevel := new(zapcore.Level)
	if err := logLevel.UnmarshalText([]byte(cfg.Log.Level)); err != nil {
		*logLevel = zapcore.InfoLevel
	}

	core := zapcore.NewCore(encoder, writeSyncer, logLevel)

	// zap.AddCaller() 会在日志中加入调用函数的文件名和行号
	// zap.AddCallerSkip(1) 会跳过调用函数的文件名和行号
	// 当我们不是直接使用初始化好的logger实例记录日志，而是将其包装成一个函数等，此时日录日志的函数调用链会增加，想要获得准确的调用信息就需要通过AddCallerSkip函数来跳过
	logger := zap.New(
		core,
		zap.AddCaller(),                   // 调用文件和行号，内部使用 runtime.Caller
		zap.AddCallerSkip(1),              // 封装了一层，调用文件去除一层(runtime.Caller(1))
		zap.AddStacktrace(zap.ErrorLevel), // 输出调用堆栈，Error 时才会显示 stacktrace
	)

	// 替换全局的 logger
	zap.ReplaceGlobals(logger)

	return logger
}

func getLogWriter(cfg *configs.Config) zapcore.WriteSyncer {
	// 因为 zap 本身不支持切割归档日志文件，因此需要借助第三方库 lumberjack 来实现
	// lumberjack.Logger 实现了 io.WriteSyncer 接口，可以直接作为 zap 的 WriteSyncer 使用
	// lumberjack.Logger 的参数如下：
	// Filename 日志文件的位置
	// MaxSize 每个日志文件保存的最大尺寸 单位：MB
	// MaxBackups 保留旧文件的最大个数
	// MaxAge 保留旧文件的最大天数
	// Compress 是否压缩
	// LocalTime 是否使用本地时间
	lumberJackLogger := &lumberjack.Logger{
		Filename:   filepath.Join(cfg.Log.LogDir, cfg.Log.LogFileName),
		MaxSize:    cfg.Log.MaxSize,
		MaxBackups: cfg.Log.MaxBackups,
		MaxAge:     cfg.Log.MaxAge,
		Compress:   cfg.Log.Compress,
		LocalTime:  cfg.Log.LocalTime,
	}

	var writeSyncer zapcore.WriteSyncer
	if cfg.App.Mode == "dev" {
		// 开发模式下，日志输出到控制台，同时也输出到文件
		writeSyncer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(lumberJackLogger))
	} else {
		writeSyncer = zapcore.AddSync(lumberJackLogger)
	}

	return writeSyncer
}

func getEncoder(mode string) zapcore.Encoder {
	var encoderConfig zapcore.EncoderConfig
	if mode == "dev" {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
		// 大小写编码器（关键词高亮）
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		encoderConfig = zap.NewProductionEncoderConfig()
		// 大小写编码器
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	}

	// 时间格式
	encoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder

	if mode == "dev" {
		// 使用 Console 格式输出日志
		return zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 使用 JSON 格式输出日志
	return zapcore.NewJSONEncoder(encoderConfig)
}
