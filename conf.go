package conf

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"
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

func init() {
	//由于 init 方法的执行顺序问题，如果日志尚未初始化，要先初始化
	if Log == nil {
		logInit()
	}
	newConf()
}

//confKeys 自定义类型key
type confKeys []string

//String 返回key字符串
func (k confKeys) toString() string {
	return strings.Join(k, ".")
}

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

type conf struct {
	//配置数据
	data map[string]map[string]result
	//当前选定的读取索引前缀
	indexPrefix string
	//写锁定，后续如果加入热更新写的时候不允许读操作避免大并发情况下读取到不完整数据
	mutex *sync.RWMutex
}
type result struct {
	dataType confType
	value    interface{}
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
	C.data = make(map[string]map[string]result)
	C.mutex = new(sync.RWMutex)
	//err = analysisLocalConfFile(e.Base.DataDir+"/app.toml", C)
	err = analysisXdiamondConf("", C)
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
	kvMap := make(map[string]result)
	_ = setKvMap(tmp, make(confKeys, 0), kvMap)
	c.data[file] = kvMap
	return nil
}

//解析配置中心数据
func analysisXdiamondConf(bojectName string, c *conf) error {
	data, err := ioutil.ReadFile(e.Base.DataDir + "/xdaimond.json")
	if err != nil {
		return err
	}
	var tmp interface{}
	err = json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	tmpSlice, ok := tmp.([]interface{})
	if !ok {
		return errors.New("配置中心:类型断言失败,请检查json结构" + fmt.Sprintf("%v", tmp))
	}
	for _, v := range tmpSlice {
		val, ok := v.(map[string]interface{})
		if ok {
			conf, ok := val["config"]
			if ok {
				fmt.Println(conf)
			}
		}
	}
	return nil
}
func getValue(key string, m interface{}) (interface{}, bool) {
	val, ok := m.(map[string]interface{})
	if ok {
		v, ok := val[key]
		if ok {
			return v, true
		}
	}
	return nil, false
}

//形成kv结构
func setKvMap(m interface{}, keys confKeys, kvMap map[string]result) error {
	tmp, ok := m.(map[string]interface{})
	if !ok {
		return errors.New("配置文件:类型断言失败,请检查配置文件内容" + fmt.Sprintf("%v", tmp))
	}
	for k, v := range tmp {
		keyNodes := append(keys, k)
		switch v.(type) {
		case map[string]interface{}:
			_ = setKvMap(v, keyNodes, kvMap)
		case []interface{}:
			kvMap[keyNodes.toString()] = result{Array, v}
		case string:
			kvMap[keyNodes.toString()] = result{String, v}
		case int, int8, int16, int32, int64:
			kvMap[keyNodes.toString()] = result{Int, v}
		case uint, uint8, uint16, uint32, uint64:
			kvMap[keyNodes.toString()] = result{Uint, v}
		case float32, float64:
			kvMap[keyNodes.toString()] = result{Float, v}
		case bool:
			kvMap[keyNodes.toString()] = result{Bool, v}
		case time.Time:
			kvMap[keyNodes.toString()] = result{Time, v}
		default:
			kvMap[keyNodes.toString()] = result{Undefined, v}
		}
	}
	return nil
}

//T 测试
func T() {
	fmt.Println("测试")
}
