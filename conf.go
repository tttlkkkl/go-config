package conf

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

const (
	format = "json"
	uri    = "/clientapi/config"
)

//C 全局变量，供外部调用
var C *conf
var e *env

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

func init() {
	//由于 init 方法的执行顺序问题，如果日志尚未初始化，要先初始化
	if Log == nil {
		logInit()
	}
	newConf()
}

//初始化
func newConf() {
	envFile, config := getOption()
	fmt.Println(envFile, config)
	var err error
	e, err = getEnv(envFile)
	fmt.Println(e)
	if err != nil {
		Log.Fatal("初始化配置环境失败", err)
	}
	C = new(conf)
	err = analysisLocalConfFile(e.Base.DataDir+"/app.toml", C)
	fmt.Println(err)
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
		Log.Fatal("日志文件解析失败:", err)
	}
	for k, v := range tmp {
		fmt.Printf("k:%s v:%s \n", k, v)
	}
	fmt.Printf("类型:%T ,值:%s\n", getTypeOf(c.data), getTypeOf(c.data))
	fmt.Println("数值", tmp)
	return nil
}

//T 测试
func T() {
	fmt.Println("测试")
}
