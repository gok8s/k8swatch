package zlog

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	"sync"

	"gopkg.in/natefinch/lumberjack.v2"

	"go.uber.org/zap"

	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger
var Slogger *zap.SugaredLogger
var Atom zap.AtomicLevel
var initialized uint32
var mux sync.Mutex

type Field = zapcore.Field

func init() {
	if Logger == nil || Slogger == nil {
		//fmt.Println("initlogger...")
		GetInstance(true, false, "", "debug", "json")
		//newLogger(true, false, "", "debug", "json")
		defer Logger.Sync()
	}
}

func GetInstanceV2(isStdOut, isSave bool, logfile, level, encodingType string) *zap.Logger {
	fmt.Println("Starting GetInstancd")
	if atomic.LoadUint32(&initialized) == 1 {
		fmt.Println("Logger has been created, return directly...")
		return Logger
	}
	mux.Lock()
	defer mux.Unlock()
	if Logger == nil {
		Logger = newLogger(isStdOut, isSave, logfile, level, encodingType)
		fmt.Println("Logger is nil,Creating new Logger...")
		atomic.StoreUint32(&initialized, 1)
	} else {
		fmt.Printf("Logger is not nil:%+v\n", Logger)
	}
	return Logger
}

func GetInstance(isStdOut, isSave bool, logfile, level, encodingType string) *zap.Logger {
	var once sync.Once
	once.Do(func() {
		Logger = newLogger(isStdOut, isSave, logfile, level, encodingType)
	})
	return Logger
}

func newLogger(isStdOut, isSave bool, logfile, level, encodingType string) *zap.Logger {
	// 保存日志路径
	if isSave && logfile == "" {
		logfile = "out.log" // 默认
	}

	group := os.Getenv("productName")
	env := os.Getenv("env")
	serviceName := os.Getenv("serviceName")
	podName := os.Getenv("MY_POD_NAME")

	// 设置日志级别
	var logLevel zapcore.Level
	switch strings.ToUpper(level) {
	case "INFO":
		logLevel = zap.InfoLevel
	case "DEBUG":
		logLevel = zap.DebugLevel
	case "ERROR":
		logLevel = zap.ErrorLevel
	case "WARN":
		logLevel = zap.WarnLevel
	case "PANIC":
		logLevel = zap.PanicLevel
	case "FATAL":
		logLevel = zap.FatalLevel
	}

	Atom = zap.NewAtomicLevelAt(logLevel)
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:    "timeLocal", //graylog standard
		LevelKey:   "logLevel",  //graylog standard
		NameKey:    "Logger",
		CallerKey:  "caller",
		MessageKey: "short_message", //graylog standard
		//StacktraceKey:  "stacktrace",
		LineEnding:  zapcore.DefaultLineEnding,
		EncodeLevel: zapcore.LowercaseLevelEncoder, // 小写编码器
		//EncodeTime:  zapcore.EpochTimeEncoder,
		EncodeTime:     TimeEncoder, //graylog standard
		EncodeDuration: zapcore.NanosDurationEncoder,
		//	EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller: zapcore.ShortCallerEncoder, //段路径编码器,如果callerKey定义了，但EncodeCaller没定义，会导致log初始化失败并panic
		//EncodeCaller:   zapcore.FullCallerEncoder, // 全路径编码器
	}

	var encoder zapcore.Encoder
	switch encodingType {
	case "json":
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	case "console":
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	default:
		fmt.Println("encodingType undefined")
	}

	config := zap.Config{
		//如下参数都在config.build()生效，如不适用config.build，则需自己实现。Sampling和InitialFields已自己实现
		Level:         Atom,          // 日志级别
		Development:   false,         // 开发模式，堆栈跟踪
		Encoding:      encodingType,  // 输出格式 console 或 json
		EncoderConfig: encoderConfig, // 编码器配置

		/*	Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},*/
		InitialFields: map[string]interface{}{
			"serviceName": serviceName,
			"group":       group,
			"env":         env,
			"podName":     podName}, // 初始化字段，如：添加一个服务器名称

	}

	var writers []zapcore.WriteSyncer

	if isStdOut {
		writers = append(writers, os.Stdout)
		config.OutputPaths = append(config.OutputPaths, "stdout")
		config.ErrorOutputPaths = []string{"stderr"}
	}

	if isSave && logfile != "" {
		lumlog := &lumberjack.Logger{
			Filename:   logfile,
			MaxSize:    1024, // megabytes //1MB*1024=1GB
			MaxBackups: 10,
			MaxAge:     7, // days
			LocalTime:  true,
		}
		writers = append(writers, zapcore.AddSync(lumlog))
		config.OutputPaths = append(config.OutputPaths, logfile)
		config.ErrorOutputPaths = append(config.ErrorOutputPaths, logfile)
	}

	output := zapcore.NewMultiWriteSyncer(writers...)

	opts := []zap.Option{zap.AddCallerSkip(1)}
	opts = append(opts, zap.AddCaller()) //启用caller

	if len(config.InitialFields) > 0 {
		fs := make([]Field, 0, len(config.InitialFields))
		keys := make([]string, 0, len(config.InitialFields))
		for k := range config.InitialFields {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fs = append(fs, zap.Any(k, config.InitialFields[k]))
		}
		opts = append(opts, zap.Fields(fs...))
	}

	if config.Sampling != nil {
		opts = append(opts, zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewSampler(core, time.Second, int(config.Sampling.Initial), int(config.Sampling.Thereafter))
		}))
	}
	//Logger, _ = zap.NewProductionConfig().Build() //这是基于官方写好的生产配置生成的实例
	//Logger = zap.New(zapcore.NewCore(encoder, &Discarder{}, Atom)) //这是将输出忽略，不产生IO
	Logger = zap.New(zapcore.NewCore(encoder, output, Atom), opts...)

	//Logger.Info("zlogger 初始化成功")
	Slogger = Logger.Sugar()
	return Logger
}

func TimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000 +0800"))
}

func Info(s string, args ...Field) {
	Logger.Info(s, args...)
}

func Infof(s string, args ...interface{}) {
	Slogger.Infof(s, args...)
}

func Debug(s string, args ...Field) {
	Logger.Debug(s, args...)
}
func Debugf(s string, args ...interface{}) {
	Slogger.Debugf(s, args...)
}

func Warn(s string, args ...Field) {
	Logger.Warn(s, args...)
}
func Warnf(s string, args ...interface{}) {
	Slogger.Warnf(s, args...)
}

func Error(s string, args ...Field) {
	Logger.Error(s, args...)
}
func Errorf(s string, args ...interface{}) {
	Slogger.Errorf(s, args...)
}

func Panic(s string, args ...Field) {
	Logger.Panic(s, args...)
}
func Panicf(s string, args ...interface{}) {
	Slogger.Panicf(s, args...)
}

func Fatal(s string, args ...Field) {
	Logger.Fatal(s, args...)
}
func Fatalf(s string, args ...interface{}) {
	Slogger.Fatalf(s, args...)
}
