#### 使用之前请移步:[toml规范](https://github.com/toml-lang/toml/blob/master/versions/cn/toml-v0.4.0.md)

#### 为规范和实现golang微服务配置统一管理，建立本类库。本类库支持本地toml格式的配置文件加载、xdiamond的HTTP接入和xdiamond的TCP接入。首要支持本地配置文件读取和解析。

##### 几个可能需要理解的概念:
- 配置环境:定义配置环境,比如`dev、test、product`。默认值为dev。
- 配置路径:本地配置集中存储路径，所有配置都将在此路径下查找。配置层级组织规则为:配置目录+配置环境名称。公共配置路径固定为"comm"。具体见后续示例。Windows默认值:`C:/web_go_config`,其他系统默认值:`/var/web_go_config`。
- 环境变量:通过定义环境变量`WEB_GO_CONFIG_ENV`来指定当前机器配置环境,通过定义环境变量`WEB_GO_CONFIG_PATH`来指定当前机器配置路径。
- 配置对象:描述一个配置文件。一个配置文件解析之后生成一个配置对象，对于xdiamond配置中心而言一个project下的配置版本对应一个配置对象。
- 配置标志：唯一描述一个配置对象的字符串。对本一个地配置文件来说配置标志=相对目录名+"."+文件名(不包括后缀),对配置中心而言配置标志=project+"."+version。
- 配置来源:标识一个配置对象来源。通过内置常量来定义，目前有本地文件、配置中心HTTP、配置中心TCP、本地备份（这个暂时是不受支持的）。
- 结果对象:描述一个配置值。结果对象提供基本类型转换。
- 所有配置最终都以kv形式获取，用"."来
##### 使用(请确保已引入conf包,以下说明均基于假设当前配置环境为`dev`):
###### 本地配置文件支持(假设当前配置路径为`/var/web_go_config`):
- 配置文件格式:toml。使用之前请移步:[toml规范](https://github.com/toml-lang/toml/blob/master/versions/cn/toml-v0.4.0.md)。
- 读取公共配置目录下的`app.toml`(此时文件完整存储路径应为：`/var/web_go_config/dev/comm/app.toml`):`c := conf.NewConfig("comm.app", conf.SourceFile)`
###### 启用配置中心必须在本地公共配置目录下面建立名为`xdiamond.toml`的配置文件，以指定配置中心服务器地址以及授权信息，配置内容见类库目录`_examples/dev/comm/xdiamond.toml`
###### HTTP方式加载配置中心配置

- 读取项目`crm`下的版本为`1.0.1`的配置:`c := conf.NewConfig("crm.1.0.1", conf.SourceXdaHTTP)`

###### TCP方式加载配置中心配置:

- 同样读取项目`crm`下的版本为`1.0.1`的配置:`c := conf.NewConfig("crm.1.0.1", conf.SourceXdaTCP)`

- 首次读取会先启动TCP同步客户端并拉取配置内容，配置中心回推配置内容之后实例化函数才会返回。

- 异步回调，通过`func SetCallbackFunc(handel CallbackHandel)`可以设置回调函数，当配置中心配置变更时会回调此方法。

- 断线重连支持:重连尝试次数20次，每次间隔5秒。

##### 方法说明:
- `func DisableCache()`:禁止在内存中缓冲配置数据,默认情况下会在内存中留存一份配置数据，重复读取时将不再读取文件或者HTTP配置中心,对于TCP配置中心此方法无效

- `func SetCallbackFunc(handel CallbackHandel)`:设置回调函数,配置文件解析完毕时尝试调用此方法。 

- 回调方法需要实现以下接口:
```golang
type CallbackHandel interface {
    CallbackHandel(fileName string, co *ConfigObject)
}
```
- 配置对象结构体:
```golang
type ConfigObject struct {
    
}
```
- `func NewConfig(fileName string, source Source) *ConfigObject`:实例化一个配置对象

- `func (c *ConfigObject) All() map[string]Result`:获取一个配置对象全部配置

- `func (c *ConfigObject) Exists() bool`:判断配置对象是否成功加载一个配置文件，一般而言如果配置信息不存在需要中断程序，是否保留此方法需进一步商榷。

- `func (c *ConfigObject) Get(key string) *Result`:获取一个结果对象，可以基于此对象提供的方法直接获取一些基本类型的值

- 配置值结构体:
```golang
type Result struct {
}
```
- `func (r *Result) Default(defaultValue interface{}) *Result`:设置一个默认值，当配置值不存在时将返回此配置值的一个配置值对象。

- `func (r *Result) Bool() bool`:以布尔型返回配置值

- `func (r *Result) Exists() bool`:判断配置文件中是否提供这个配置值

- `func (r *Result) Float() float64`:以浮点类型返回配置值

- `func (r *Result) Int() int64`:以int64类型返回配置值

- `func (r *Result) Slice() []interface{}`:Slice 以切片返回配置值在toml中 类似 k=[1,2]这样配置会以切片返回(xdiamond配置中心不支持这种配置)，除此之外返回包含配置值的切片

- `func (r *Result) SliceMap() []map[string]interface{}`:SliceMap 断言返回类似以下配置:

```toml
[[crm.slave]]
    addr = "localhost:6379"
    password = ""
    db = 0
[[crm.slave]]
    addr = "localhost:6379"
    password = ""
    db = 0
```
- `func (r *Result) String() string`:以字符串类型返回配置值

- `func (r *Result) Time() time.Time`:以时间类型返回配置值，时间格式依照toml以RFC3339因特网标准时间为准

- `func (r *Result) ToDateTime() string`:尝试以Y-m-d h:i:s的格式返回时间配置值字符串

- `func (r *Result) Uint() uint64`:以uint64返回配置值，注意：负数转无符号数得到的值可能不是你预期的

- `func (r *Result) Value() interface{}`:返回原解析配置值而不进行任何转化,需要基于结果进行断言和转换处理

- `type Source int`:配置来源

- 配置来源常量:

```golang
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
```

- 简单使用示例:
假设名为"app.toml"的配置文件位于公共配置目录，其配置内容如下:
```toml
path="/etc/config"
[base]
  name = "Tom Preston-Werner"
  organization = "GitHub"
  bio = "GitHub Cofounder & CEO\nLikes tater tots and beer."
  dob = 2018-05-27T07:32:00Z # RFC3339因特网标准时间
  int = 1
  float =1.1
  bool = true
[[default.master]]
	addr = "localhost:6379"
	password = ""
	db = 0
[[default.slave]]
	addr = "localhost:6379"
    password = ""
    db = 0
[[default.slave]]
	addr = "localhost:6379"
    password = ""
    db = 0
```
示例(注意:包的引入仅作为参考，请安实际进行调整):

```golang
package main

import (
	"conf"
	"fmt"
)

func main() {
	//实例化一个配置对象
	//配置中心
	// c := conf.NewConfig("golang-test.1.0", conf.SourceXdaHTTP)

	//设置一个配置变更回调
	callback := new(configChange)
	conf.SetCallbackFunc(callback)
	c := conf.NewConfig("comm.app", conf.SourceFile)
	//取得path的值，设默认值
	a1 := c.Get("path").Default("/www/config").String()
	fmt.Println("path:", a1)
	//取得base配置节下的int的值，设默认值
	a2 := c.Get("base.int").Default(1).Int()
	fmt.Println("base.int:", a2)
	//取得base配置节下dob的值，不设默认值
	a3 := c.Get("base.dob").Time()
	fmt.Println("base.dob:", a3)
	//判断一个配置值是否存在
	if !c.Get("no").Exists() {
		fmt.Println("配置no不存在!")
	}
	if c.Get("default.master").Exists() {
		fmt.Println("配置default.master存在!")
	}
	//获取嵌套表default.slave内容
	a4 := c.Get("default.slave").SliceMap()
	fmt.Println("default.slave:", a4)
	fmt.Println("default.slave子节点数量:", len(a4))
	//或者可以这样做--举一反三，如果提供的基本类型转换无法满足需求时可以获取原始值自行断言处理
	a5 := c.Get("default.slave").Value()
	a5t, ok := a5.([]map[string]interface{})
	if ok {
		fmt.Println("default.slave:", a5t)
		fmt.Println("default.slave子节点数量:", len(a5t))
	}

}

type configChange struct{}

func (c *configChange) CallbackHandel(fileName string, co *conf.ConfigObject) {
	fmt.Println("配置对象：", fileName, "有变更...")
	//打印全部配置内容
	for k, v := range co.All() {
		fmt.Printf("key:%v , value:%v", k, v.Value())
	}
}
```
