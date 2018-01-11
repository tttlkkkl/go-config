//Package conf golang微服务配置管理组件，以toml为配置格式，同时接入xdiamond配置中心。力求实现微服务配置集中管理。
package conf

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
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
var c *conf
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

//配置数据存储结构
type conf struct {
	//配置数据
	data map[string]ConfigObject
	//写锁定，后续如果加入热更新写的时候不允许读操作避免大并发情况下读取到不完整数据
	mutex *sync.RWMutex
}

//ConfigObject 配置对象
type ConfigObject struct {
	data map[string]Result
	//当前选定的读取索引前缀
	indexPrefix string
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

//C 获取一个配置对象
func C(objectName string) *ConfigObject {
	object, ok := c.data[objectName]
	if ok {
		return &object
	}
	return new(ConfigObject)
}

//Get 获取一个配置结果
func (c *ConfigObject) Get(key string) *Result {
	if key == "" {
		return new(Result)
	}
	key = c.indexPrefix + key
	r, ok := c.data[key]
	if ok {
		return &r
	}
	return new(Result)
}

//All 获取全部配置
func (c *ConfigObject) All() map[string]Result {
	return c.data
}

//Exists 配置对象是否是真的从配置数据中解析得来的
func (c *ConfigObject) Exists() bool {
	return c.isExistence
}

//SetPrefix 设置配置读取索引前缀，此方法有助于简短优雅的读取一个配置节信息,设置成功后每次使用Get函数都会自动在key前连接这个prefix
//为了避免某些并发情况下用此方法读取配置时出现混乱，必须在清除后才能重新设置
func (c *ConfigObject) SetPrefix(prefix string) bool {
	if c.indexPrefix != "" {
		c.indexPrefix = prefix
		return true
	}
	return false
}

//UnsetPrefix 清除已设置的key前缀
func (c *ConfigObject) UnsetPrefix() {
	c.indexPrefix = ""
}

//Exists 判断配置值是否是真实从配置文件中解析得来的
func (r *Result) Exists() bool {
	return r.isExistence
}

//Value 返回原解析配置值而不进行任何转化
func (r *Result) Value() interface{} {
	return r.Value
}

//String 以字符串返回配置值
func (r *Result) String() string {
	return fmt.Sprintf("%v", r.value)
}

//Slice 以切片返回配置值在toml中 类似 k=[1,2]这样配置会以切片返回(xdiamond 配置中心不支持这种配置)，除此之外返回包含配置值的切片
func (r *Result) Slice() []interface{} {
	if r.dataType == Array {
		v, ok := r.value.([]interface{})
		if !ok {
			return make([]interface{}, 0, 0)
		}
		return v
	}
	v := make([]interface{}, 0)
	return append(v, r.value)
}

//Time 以时间格式返回配置值，时间格式依照toml以RFC3339因特网标准时间为准
func (r *Result) Time() time.Time {
	if r.dataType == Time {
		v, ok := r.value.(time.Time)
		if ok {
			return v
		}
	}
	if r.dataType == String {
		v, ok := r.value.(string)
		if ok {
			t, err := time.Parse(time.RFC3339, v)
			if err == nil {
				return t
			}
		}
	}
	return time.Unix(0, 0)
}

//ToDateTime 尝试以Y-m-d h:i:s的格式返回时间配置值字符串
func (r *Result) ToDateTime() string {
	timeLayout := "2006-01-02 15:04:05"
	return r.Time().Format(timeLayout)
}

//Bool 已布尔型返回配置值
func (r *Result) Bool() bool {
	switch r.dataType {
	case String:
		v, ok := r.value.(string)
		if !ok {
			return false
		}
		n, err := strconv.ParseBool(v)
		if err != nil {
			return false
		}
		return n
	case Int, Uint, Float:
		v, ok := r.value.(float64)
		if !ok {
			return false
		}
		return v != 0
	case Bool:
		v, ok := r.value.(bool)
		if !ok {
			return false
		}
		return v
	case Array:
		v, ok := r.value.([]interface{})
		if !ok {
			return false
		}
		return len(v) > 0
	case Time:
		v, ok := r.value.(time.Time)
		if !ok {
			return false
		}
		return v.Unix() > 0
	case Undefined:
		return !(r.value == nil)
	default:
		return false
	}
}

//Float 以浮点类型返回配置值
func (r *Result) Float() float64 {
	switch r.dataType {
	case String:
		v, ok := r.value.(string)
		if !ok {
			return float64(0)
		}
		n, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return float64(0)
		}
		return n
	case Int, Uint, Float:
		v, ok := r.value.(float64)
		if !ok {
			return float64(0)
		}
		return float64(v)
	case Bool:
		v, ok := r.value.(bool)
		if !ok {
			return float64(0)
		}
		if v {
			return float64(1)
		}
	default:
		return float64(0)
	}
	return float64(0)
}

//Int 已int64类型返回配置值
func (r *Result) Int() int64 {
	switch r.dataType {
	case String:
		v, ok := r.value.(string)
		if !ok {
			return int64(0)
		}
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return int64(0)
		}
		return n
	case Int, Uint, Float:
		v, ok := r.value.(int64)
		if !ok {
			return int64(0)
		}
		return int64(v)
	case Bool:
		v, ok := r.value.(bool)
		if !ok {
			return int64(0)
		}
		if v {
			return int64(1)
		}
	default:
		return int64(0)
	}
	return int64(0)
}

//Uint 以uint64返回配置值，注意：负数转无符号数得到的值可能不是你预期的
func (r *Result) Uint() uint64 {
	return uint64(r.Int())
}

//初始化
func newConf() {
	envFile, _ := getOption()
	var err error
	e, err = getEnv(envFile)
	if err != nil {
		Log.Fatal("初始化配置环境失败", err)
	}
	c = new(conf)
	c.data = make(map[string]ConfigObject)
	c.mutex = new(sync.RWMutex)
	err = analysisLocalConfFile(e.Base.DataDir+"/app.toml", c)
	if err != nil {
		Log.Fatal("配置文件"+e.Xdiamond.ArtifactID+"解析失败", err)
	}
	err = analysisXdiamondConf(e.Xdiamond.ArtifactID, c)
	if err != nil {
		Log.Fatal("配置中心"+e.Xdiamond.ArtifactID+"解析失败", err)
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
	kvMap := make(map[string]Result)
	err = setKvMap(tmp, make(confKeys, 0), kvMap)
	if err != nil {
		return err
	}
	c.data[file] = ConfigObject{kvMap, "", true}
	return nil
}

//解析配置中心数据
func analysisXdiamondConf(objectName string, c *conf) error {
	data, err := ioutil.ReadFile(e.Base.DataDir + "/xdaimond.json")
	if err != nil {
		return err
	}
	if objectName == "" {
		return errors.New("无效的配置中心对象名称")
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
	var kvMapTmp = make(map[string]interface{})
	for _, v := range tmpSlice {
		val, ok := getValue("config", v)
		if !ok {
			continue
		}
		key, keyOk := getValue("key", val)
		value, valueOk := getValue("value", val)
		if keyOk && valueOk {
			key, ok := key.(string)
			if !ok {
				continue
			}
			kvMapTmp[key] = value
		}
	}
	kvMap := make(map[string]Result)
	err = setKvMap(kvMapTmp, make(confKeys, 0), kvMap)
	if err != nil {
		return err
	}
	c.data[objectName] = ConfigObject{kvMap, "", true}
	return nil
}

//从接口类型的map里面获取指定数据
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
		case []interface{}:
			kvMap[keyNodes.toString()] = Result{Array, v, true}
		case string:
			kvMap[keyNodes.toString()] = Result{String, v, true}
		case int, int8, int16, int32, int64:
			kvMap[keyNodes.toString()] = Result{Int, v, true}
		case uint, uint8, uint16, uint32, uint64:
			kvMap[keyNodes.toString()] = Result{Uint, v, true}
		case float32, float64:
			kvMap[keyNodes.toString()] = Result{Float, v, true}
		case bool:
			kvMap[keyNodes.toString()] = Result{Bool, v, true}
		case time.Time:
			kvMap[keyNodes.toString()] = Result{Time, v, true}
		default:
			kvMap[keyNodes.toString()] = Result{Undefined, v, true}
		}
	}
	return nil
}
