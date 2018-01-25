// Package conf golang微服务配置管理组件，以toml为配置格式，同时接入xdiamond配置中心。力求实现微服务配置集中管理。
package conf

import (
	"sync"
)

// Type 自定义类型，用于归纳基本数据类型
type confType int

const (
	//String 字符串类型
	String confType = iota
	//Int 有符号整型 int32 也表示Unicode码点即rune类型
	Int
	//Uint 无符号整型uint8也表示字符型 byte类型
	Uint
	//Float 浮点数
	Float
	//Bool 布尔类型
	Bool
	//Array 数组  []interface{}
	Array
	//Time 时间类型
	Time
	//Undefined 未定义的类型
	Undefined
)

// Source 配置来源
type Source int

const (
	// SourceFile 配置来源，文件
	SourceFile Source = iota + 1
	// SourceXdaimond 配置来源，配置中心
	SourceXdaimond
	// SourceBackups 配置来源，本地备份
	SourceBackups
)

//配置数据存储结构
type conf struct {
	//配置数据
	data map[string]ConfigObject
	//存储配置来源  1 本地配置文件   2 配置中心
	source map[string]Source
	//写锁定，后续如果加入热更新写的时候不允许读操作避免大并发情况下读取到不完整数据
	mutex *sync.RWMutex
}

//ConfigObject 配置对象
type ConfigObject struct {
	data map[string]Result
	//标记是否存在
	isExistence bool
}

//Result 配置数据解析结果
type Result struct {
	dataType confType
	value    interface{}
	//标记是否存在
	isExistence bool
}

// Handel 当配置有更新时调用此方法
type Handel interface {
	handel(co *ConfigObject)
}

//解析统一接口
type analysis interface {
	analysisConfig(fileName string) (map[string]interface{}, error)
}

func C() {
	println("xxx")
}
