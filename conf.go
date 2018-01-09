package conf

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

const (
	format = "json"
	uri    = "/clientapi/config"
)

//C 全局变量，供外部调用
var C *conf
var e *env

// Type 自定义类型，用于归纳基本数据类型
type Type int

const (
	//String 字符串类型
	String Type = iota
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
)
func init() {
	//由于 init 方法的执行顺序问题，如果日志尚未初始化，要先初始化
	if Log == nil {
		logInit()
	}
	newConf()
}
//Key 自定义类型key
type Key string
func 
//configCenter xdiamond配置中心连接定义
type xdiamond struct {
	//groupId 对应groupId
	GroupID string `toml:"group_id"`
	//artifactId 对应project
	ArtifactID string `toml:"artifact_id"`
	//string 对应环境
	Profile string
	//version 配置版本,指project的版本
	Version string
	//secretKey auth认证key
	SecretKey string `toml:"secret_key"`
	//address 配置中心服务器地址
	Address string
}

// 环境定义
type env struct {
	Xdiamond xdiamond
	Base     base
}

// 基础配置
type base struct {
	//数据路径
	DataDir string `toml:"data_dir"`
	//环境
	Env string
	//1 文件配置 2 配置中心 3配置中心缓存
	First   int
	Options envOption
}

//基本环境配置可选项
type envOption struct {
	//cachePath 配置缓存路径
	CachePath string `toml:"cache_path"`
	//日志路径
	LogPath string `toml:"log_tath"`
	//confPath 默认配置路径
	ConfPath string `toml:"config_path"`
}

//ds 数据结构类型 data structure
type ds map[string]interface{}

type conf struct {
	//配置数据 第一层key为配置文件名 如果是配置中心xdiamond则为project名称
	data ds
	//存储变量类型
	types ds
}

//初始化
func newConf() {
	envFile, _ := getOption()
	var err error
	e, err = getEnv(envFile)
	if err != nil {
		Log.Fatal("初始化配置环境失败", err)
	}
	C = new(conf)
	err = analysisLocalConfFile(e.Base.DataDir+"/app.toml", C)
	if err != nil {
		Log.Fatal("配置解析失败", err)
	}
}

//获取环境配置
func getEnv(envFile string) (*env, error) {
	env := new(env)
	_, err := toml.DecodeFile(envFile, env)
	return env, err
}

//解析 toml日志文件
func analysisLocalConfFile(confFile string, c *conf) error {
	var tmp map[string]interface{}
	_, err := toml.DecodeFile(confFile, &tmp)
	if err != nil {
		Log.Fatal("配置文件"+confFile+"解析失败:", err)
	}
	_, file := filepath.Split(confFile)
	file = strings.TrimSuffix(file, filepath.Ext(file))
	if file == "" {
		return errors.New("无效的配置文件名称")
	}
	_ = c.fillMap(tmp, file)
	return nil
}

//填充存储结构
func (c *conf) fillMap(m interface{}, objectName string,key string) error {
	tmp, ok := m.(map[string]interface{})
	if !ok {
		return errors.New("类型断言失败")
	}
	for k, v := range tmp {
		switch o := v.(type) {
		case map[string]interface{}:
			c.fillMap(v, objectName)
		case []interface{}:
			c.fillMap(v, objectName)
		case string:
			fmt.Printf("[]interface: key:%s type: %T value:%v\n", k, o, v)
		case int, int8, int16, int32, int64:
			fmt.Printf("[]interface: key:%s type: %T value:%v\n", k, o, v)
		case uint, uint8, uint16, uint32, uint64:
			fmt.Printf("[]interface: key:%s type: %T value:%v\n", k, o, v)
		case float32, float64:
			fmt.Printf("[]interface: key:%s type: %T value:%v\n", k, o, v)
		case bool:
			fmt.Printf("[]interface: key:%s type: %T value:%v\n", k, o, v)
		case time.Time:
			fmt.Printf("[]interface: key:%s type: %T value:%v\n", k, o, v)
		default:
			fmt.Printf("未知类型: key:%s type: %T value:%v\n", k, o, v)
		}
	}

	return nil
}

//T 测试
func T() {
	fmt.Println("测试")
}
