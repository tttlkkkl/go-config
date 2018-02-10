// Package conf golang微服务配置管理组件，以toml为配置格式，同时接入xdiamond配置中心。力求实现微服务配置集中管理。
package conf

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

//全局环境
var e *env
var c *conf

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
	// SourceFile 配置源，文件
	SourceFile Source = iota + 1
	// SourceXdaHTTP 配置源，配置中心http
	SourceXdaHTTP
	// SourceXdaTCP 配置源，配置中心TCP
	SourceXdaTCP
	// SourceBackups 配置来源，本地备份
	SourceBackups
)

//配置数据存储结构
type conf struct {
	//配置数据
	data map[string]ConfigObject
	//写锁定，后续如果加入热更新写的时候不允许读操作避免大并发情况下读取到不完整数据
	mutex *sync.RWMutex
	//是否将数据缓存在内存中
	isCache bool
	// 回调函数
	handel CallbackHandel
}

// CallbackHandel 当配置有更新时调用此方法
type CallbackHandel interface {
	CallbackHandel(fileName string, co *ConfigObject)
}

//解析统一接口
type analysis interface {
	analysisConfig(fileName string) (map[string]interface{}, error)
}

//confKeys 自定义类型key
type confKeys []string

//String 返回key字符串
func (k confKeys) toString() string {
	return strings.Join(k, ".")
}

// 初始化必要的变量
func init() {
	var err error
	logInit()
	e, err = newEnv()
	if err != nil {
		Log.Fatal(err)
	}
	c = &conf{
		data:    make(map[string]ConfigObject),
		mutex:   new(sync.RWMutex),
		isCache: true,
	}
	// 设置日志路径,在此之前打印的信息还是会输出到终端
	logConf := NewConfig("comm.log", SourceFile)
	logDir := logConf.Get("base.dir")
	dir := logDir.String()
	if logDir.Exists() && dir != "" && false {
		dirInfo, err := os.Stat(dir)
		if err != nil {
			if os.IsNotExist(err) {
				err = os.MkdirAll(dir, 0664)
				if err != nil {
					Log.Fatal("创建日志目录失败:", err)
				}
			} else {
				Log.Fatal(err)
			}
		} else if !dirInfo.IsDir() {
			Log.Fatal(dir, " : 不是一个有效的目录")
		}
		logFileName := dir + "/conf.log"
		fileInfo, err := os.OpenFile(logFileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0664)
		if err != nil {
			Log.Error("日志文件打开失败...", err)
		}
		SetLogOutput(fileInfo, All)
	}
}

// NewConfig 实例化一个配置对象
func NewConfig(fileName string, source Source) *ConfigObject {
	switch source {
	case SourceFile:
		return c.getConfigObject(fileName, source, newLocalFile())
	case SourceXdaHTTP:
		return c.getConfigObject(fileName, source, newXdiamondHTTP())
	case SourceXdaTCP:
		return c.getConfigObject(fileName, source, newXdiamondTCP())
	case SourceBackups:
		return new(ConfigObject)
	}
	return new(ConfigObject)
}

// DisableCache 禁止在内存中缓冲配置数据
func DisableCache() {
	c.isCache = false
}

// SetCallbackFunc 设置回调函数
func SetCallbackFunc(handel CallbackHandel) {
	c.handel = handel
}

// getConfigObject 获取一个配置对象
func (c *conf) getConfigObject(fileName string, source Source, obj analysis) *ConfigObject {
	if c.isCache || source == SourceXdaTCP {
		object, ok := c.data[fileName]
		if ok {
			return &object
		}
	}
	tmp, err := obj.analysisConfig(fileName)
	if err != nil {
		//尝试从备份文件读取
		if source == SourceXdaHTTP || source == SourceXdaTCP {
			Log.Warning("配置中心连接失败..." + err.Error())
			Log.Info("尝试从本地备份读取配置...")
			tmps, err := backupRecovery(fileName)
			if err != nil {
				Log.Fatal(err)
			}
			return c.genConfigObject(fileName, source, tmps)
		}
		Log.Fatal(err)
	}
	return c.genConfigObject(fileName, source, tmp)
}

// 生成配置对象
func (c *conf) genConfigObject(fileName string, source Source, confMap map[string]interface{}) *ConfigObject {
	kvMap := make(map[string]Result)
	err := setKvMap(confMap, make(confKeys, 0), kvMap)
	if err != nil {
		Log.Fatal(err)
	}
	co := ConfigObject{kvMap, true, source, fileName}
	if c.isCache || source == SourceXdaTCP || source == SourceXdaHTTP {
		// 配置中心数据备份
		if source == SourceXdaTCP || source == SourceXdaHTTP {
			err = backups(fileName, confMap)
			if err != nil {
				Log.Error(err)
			}
		}
	}
	c.save(fileName, co)
	return &co
}

//  数据保存到内存
func (c *conf) save(fileName string, co ConfigObject) {
	//写锁定
	c.mutex.Lock()
	c.data[fileName] = co
	c.mutex.Unlock()
	//如果有设置回调函数，调用之
	if c.handel != nil {
		c.handel.CallbackHandel(fileName, &co)
	}
}

// setKvMap 递归设置一个kvMap
func setKvMap(m interface{}, keys confKeys, kvMap map[string]Result) error {
	tmp, ok := m.(map[string]interface{})
	if !ok {
		return errors.New("类型断言失败,配置内容格式:" + fmt.Sprintf("%v", tmp))
	}
	for k, v := range tmp {
		keyNodes := append(keys, k)
		switch v.(type) {
		case map[string]interface{}:
			_ = setKvMap(v, keyNodes, kvMap)
		default:
			kvMap[keyNodes.toString()] = genResult(v)
		}

	}
	return nil
}

// 生成结果对象
func genResult(v interface{}) Result {
	switch v.(type) {
	case []interface{}, []map[string]interface{}, [][]interface{}, [][]map[string]interface{}, map[string]interface{}:
		return Result{Array, v, true}
	case string:
		return Result{String, v, true}
	case int, int8, int16, int32, int64:
		return Result{Int, v, true}
	case uint, uint8, uint16, uint32, uint64:
		return Result{Uint, v, true}
	case float32, float64:
		return Result{Float, v, true}
	case bool:
		return Result{Bool, v, true}
	case time.Time:
		return Result{Time, v, true}
	default:
		return Result{Undefined, v, true}
	}
}
