package conf

import (
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
	error *log.Logger
	//Fatal 致命的错误日志
	fatal *log.Logger
}

func init() {
	if Log == nil {
		logInit()
	}
}

func logInit() {
	Log = new(loger)
	Log.debug = log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
	Log.info = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	Log.warning = log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	Log.error = log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	Log.fatal = log.New(os.Stdout, "FATAL: ", log.Ldate|log.Ltime|log.Lshortfile)
}

//Debug 打印调试日志
func (l *loger) Debug(args ...interface{}) {
	l.debug.Println(args...)
}

//Info 打印提示信息日志
func (l *loger) Info(args ...interface{}) {
	l.info.Fatalln(args...)
}

//Warning 打印错误日志
func (l *loger) Warning(args ...interface{}) {
	l.warning.Println(args...)
}

//Error 打印严重的错误日志
func (l *loger) Error(args ...interface{}) {
	l.error.Println(args...)
}

//Fatal 打印致命错误日志，并中断程序执行
func (l *loger) Fatal(args ...interface{}) {
	l.fatal.Fatalln(args...)
}
