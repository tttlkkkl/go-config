package conf

import (
	"errors"
	"io"
	"log"
	"os"
)

//Log 供外部使用的全局日志变量
var Log *loger

type loger struct {
	//Debug 调试日志
	debug *log.Logger
	//Info 重要提示
	info *log.Logger
	//Warning 错误日志
	warning *log.Logger
	//Error 严重的错误日志
	err *log.Logger
	//Fatal 致命的错误日志
	fatal    *log.Logger
	logLevel LogLevel
}

// LogLevel 日志级别
type LogLevel uint8

const (
	// All 全部日志
	All LogLevel = iota
	// Debug 调试日志
	Debug
	// Info 运行信息
	Info
	// Warning 需要特别注意的信息
	Warning
	// Err 错误日志
	Err
	// Fatal 致命的错误日志
	Fatal
	// Non 不打印任何日志
	Non
)

func logInit() {
	Log = new(loger)
	Log.debug = log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
	Log.info = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	Log.warning = log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	Log.err = log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	Log.fatal = log.New(os.Stdout, "FATAL: ", log.Ldate|log.Ltime|log.Lshortfile)
	Log.logLevel = 0
}

// SetLogOutput 设置日志输出
func SetLogOutput(w io.Writer, lv LogLevel) error {
	if Log == nil {
		return errors.New("日志模块未初始化")
	}
	Log.debug.SetOutput(w)
	Log.info.SetOutput(w)
	Log.warning.SetOutput(w)
	Log.err.SetOutput(w)
	Log.fatal.SetOutput(w)
	return nil
}

//Debug 打印调试日志
func (l *loger) Debug(args ...interface{}) {
	if l.logLevel > Debug {
		return
	}
	l.debug.Println(args...)
}

//Info 打印提示信息日志
func (l *loger) Info(args ...interface{}) {
	if l.logLevel > Info {
		return
	}
	l.info.Println(args...)
}

//Warning 打印错误日志
func (l *loger) Warning(args ...interface{}) {
	if l.logLevel > Warning {
		return
	}
	l.warning.Println(args...)
}

//Error 打印严重的错误日志
func (l *loger) Error(args ...interface{}) {
	if l.logLevel > Err {
		return
	}
	l.err.Println(args...)
}

//Fatal 打印致命错误日志，并中断程序执行
func (l *loger) Fatal(args ...interface{}) {
	if l.logLevel > Fatal {
		return
	}
	l.fatal.Fatalln(args...)
}
