package zlog

import (
	"fmt"
	"log"
	"os"
	"testing"

	"go.uber.org/zap/zapcore"

	"go.uber.org/zap"
)

func TestNewProduction(t *testing.T) {
	logger0 := log.New(os.Stdout, "log", 0)
	logger0.Println("logger0")

	logger1, _ := zap.NewProduction()
	logger1.Info("logger1", zap.String("name", "zs"))
	logger1.Sugar().Infof("logger1,name:%s", "zs")

	logger2, _ := zap.NewProductionConfig().Build()
	logger2.Info("logger2", zap.String("name", "zs"))

	var opts []zap.Option
	//最小配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:     "timeLocal", //graylog standard
		LevelKey:    "logLevel",  //graylog standard
		NameKey:     "Logger",
		MessageKey:  "short_message",               //graylog standard
		EncodeLevel: zapcore.LowercaseLevelEncoder, // 小写编码器
		EncodeTime:  TimeEncoder,                   //graylog standard
	}
	encoder := zapcore.NewJSONEncoder(encoderConfig)
	logger3 := zap.New(zapcore.NewCore(encoder, os.Stdout, zap.NewAtomicLevelAt(-1)), opts...)
	logger3.Info("logger3", zap.String("name", "zs"))
	//TODO 自定义默认字段，限流，日志切割，caller
}

func TestGetInstance(t *testing.T) {
	a := "debug"
	//b := 10
	GetInstance(true, false, "", "debug", "json")
	//	newLogger(true, false, "", "debug", "json")
	fmt.Println("start")
	Info("testlogxxxx", zap.String("aaaaaa", a))
	fmt.Println("end")
	ff := true
	Info("test bu", zap.Bool("yes.no", ff))
	/*SetLogger(logger)
	Info("runing loglevel")
	Infof("runing loglevel:%s age:%d", a,b)
	//Warnf()
	Info("msgxxx1",
		zap.String("name","zhangsan"),
		zap.Int("age",18),)

	logger.Info("msgxxx2",
		zap.String("name","zhangsan"),
		zap.Int("age",19),)*/
}
