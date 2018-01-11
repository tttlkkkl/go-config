## golang配置管理方案,本地toml配置文件加载,同时支持xdiamond接入，http和tcp实时同步。

## 简单使用示例
#### 引用 github.com/tttlkkkl/go-config 库编码如下
```golang
package main

import (
	"conf"
	"fmt"
)

func main() {
	//加载名为 golang-test 的配置。假设这个app就是本项目示例配置中的app.toml 要读取数据库配置节的配置可以这样做
	x := conf.C("golang-test")
	for k, v := range x.All() {
		fmt.Printf("key:%v,value:%v \n", k, v.String())
	}
	//加载名为 app 的配置
	x = conf.C("app")
	x.SetPrefix("database")
	fmt.Println(x.Get("server").String())
	fmt.Println(x.Get("ports").Slice())
	fmt.Println(x.Get("connection_max").Int())
	fmt.Println(x.Get("enabled").Bool())
	//注意一定要UnsetPrefix()否则下次再设置前缀将会失败,这可能是个糟糕的设计后期废除或者加以改进
	x.UnsetPrefix()

	//最稳定的读法是 不设置前缀，指定整个的配置值路径,像下面这样
	fmt.Println(x.Get("servers.alpha.ip").String())
    fmt.Println(x.Get("servers.alpha.dc").Value())
    
    fmt.Println(x.Get("owner.name").String())
	fmt.Println(x.Get("owner.organization").String())
	fmt.Println(x.Get("owner.dobbio").Time())
	fmt.Println(x.Get("owner.int").Int())
	fmt.Println(x.Get("owner.float").Float())
	fmt.Println(x.Get("owner.bool").Bool())

}
```

#### 运行
```golang
go run main.go -env="$GOPATH/github.com/tttlkkkl/go-config/_examples/.env.toml" -conf="xxx.toml"
```
## 注意的点
- **所有的配置在第一次加载本库时已初始化完成，不建议或者不允许将获取的配置值存储在全局变量中使用，以期为后续增加本地文件热加载，以及配置中心实时同步提供支持**
- **为了防止并发情况下新旧配置数据被混合读取而导致异常，配置文件在更新时写锁定**
- **每次获取配置前必须先获取一个配置对象，每一个配置文件对应一个配置对象**
- **对于配置中心xdiamond而言,项目名称(object)就是配置文件名称**
- **本地配置文件名为去除后缀的文件名称,当本地配置文件名为".toml"时将配置文件名置为"."**
- **xdiamond配置时间需以RFC3339因特网标准时间为准,否则Result.Time()无法取得预期的时间，此时也可以Result.Value()获取原始值并自行转换**
- **xdiamond配置中心，只支持kv配置,v只能是字符串， 不能得到如[]interface{}结构数据，如有需要自行处理**
由以上示例可以看出引用本库后程序默认支持两个参数传入,其中:
- -env 接受一个toml格式的配置文件,定义配置环境
- -conf 表示实时额外加载的配置文件或目录,如果是目录则会加载该目录下所有以".toml"为后缀的文件。

#### 环境配置完整示例:
```toml
# 基本配置
[base]
    # 数据文件存储路径
    data_dir="/mnt/share/src/conf/_examples/"
    #环境
    env="dev"
    # 优先取用何种配置,当配置中心项目名称和配置文件名称冲突时以优先加载的为准  1 文件配置 2 配置中心
    # 当本地目录有和线上配置中心项目名称(object)相同时配置中心的同步将失效
    # 建议的配置是 2
    first=1
# 以下是非必须的选项
[base.options]
    # 配置缓存路径,默认 $data_dir/cache
    cache_path="/mnt/share/src/conf/_examples/cache"
    # 日志路径 默认 $data_dir/log
    log_tath="/mnt/share/src/conf/_examples/log"
    # 配置文件存放目录，将会解析并加载该目录下的所有配置文件 默认$data_dir/conf
    config_path="/mnt/share/src/conf/_examples/conf"

# 配置中心配置
[xdiamond]
	#groupId 对应groupId
	group_id = "web"
	#artifactId 对应project 默认test
	artifact_id = "golang-test"
	#string 对应环境 这里没有配置的话会取 base.env
	profile = "dev"
	#version 配置版本,指project的版本 默认 1.0
	version ="1.0"
	#secretKey auth认证key 默认 ""
	secret_key ="68bq57jhxmi"
	#address 配置中心服务器地址 不允许添加http等前缀，ip+port  或 域名+port
	address ="10.0.200.53:8089"
    #配置中心数据同步方式,可选数值 http,tcp 。http拉取数据方式在程序启动时同步一次,默认tcp
    conn_mode="http"
```
#### go doc
```golang
PACKAGE DOCUMENTATION

package conf
    import "github.com/tttlkkkl/go-config"

    Package conf golang微服务配置管理组件，以toml为配置格式，同时接入xdiamond配置中心。力求实现微服务配置集中管理。

CONSTANTS

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

VARIABLES

var Log *loger
    Log 供外部使用的全局日志变量

TYPES

type ConfigObject struct {
    // contains filtered or unexported fields
}
    ConfigObject 配置对象

func C(objectName string) *ConfigObject
    C 获取一个配置对象

func (c *ConfigObject) All() map[string]Result
    All 获取全部配置

func (c *ConfigObject) Exists() bool
    Exists 配置对象是否是真的从配置数据中解析得来的

func (c *ConfigObject) Get(key string) *Result
    Get 获取一个配置结果

func (c *ConfigObject) SetPrefix(prefix string) bool
    SetPrefix
    设置配置读取索引前缀，此方法有助于简短优雅的读取一个配置节信息,设置成功后每次使用Get函数都会自动在key前连接这个prefix
    为了避免某些并发情况下用此方法读取配置时出现混乱，必须在清除后才能重新设置

func (c *ConfigObject) UnsetPrefix()
    UnsetPrefix 清除已设置的key前缀

type Result struct {
    // contains filtered or unexported fields
}
    Result 配置数据解析结果

func (r *Result) Bool() bool
    Bool 以布尔型返回配置值

func (r *Result) Exists() bool
    Exists 判断配置值是否是真实从配置文件中解析得来的

func (r *Result) Float() float64
    Float 以浮点类型返回配置值

func (r *Result) Int() int64
    Int 以int64类型返回配置值

func (r *Result) Slice() []interface{}
    Slice 以切片返回配置值在toml中 类似 k=[1,2]这样配置会以切片返回(xdiamond
    配置中心不支持这种配置)，除此之外返回包含配置值的切片

func (r *Result) String() string
    String 以字符串返回配置值

func (r *Result) Time() time.Time
    Time 以时间格式返回配置值，时间格式依照toml以RFC3339因特网标准时间为准

func (r *Result) ToDateTime() string
    ToDateTime 尝试以Y-m-d h:i:s的格式返回时间配置值字符串

func (r *Result) Uint() uint64
    Uint 以uint64返回配置值，注意：负数转无符号数得到的值可能不是你预期的

func (r *Result) Value() interface{}
    Value 返回原解析配置值而不进行任何转化

```