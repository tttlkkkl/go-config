package conf

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestFile(t *testing.T) {
	var err error
	err = setEnv()
	if err != nil {
		t.Error(err)
	}
	c := NewConfig("comm.app", SourceFile)
	if c.Get("title").String() != "TOML Example" {
		t.Error("字符创类型数据读取错误...")
	}
	if c.Get("base.int").Int() != 1 {
		t.Error("数字型数据读取错误...")
	}
	if ti, _ := time.Parse(time.RFC3339, "2018-05-27T07:32:00Z"); c.Get("base.dob").Time() != ti {
		t.Error("时间类型数据读取错误...")
	}
	if c.Get("base.float").Float() != float64(1.1) {
		t.Error("浮点类型数据读取错误...")
	}
	if !c.Get("base.bool").Bool() {
		t.Error("布尔类型数据读取错误...")
	}
	if c.Get("servers.alpha.ip").String() != "10.0.0.1" {
		t.Error("嵌套配置数据读取错误...")
	}
	_, ok := c.Get("clients.data").Value().([]interface{})
	if !ok {
		t.Errorf("嵌套配置数据读取错误...%T", c.Get("clients.data").Value())
	}
}
func TestXdiamondHTTP(t *testing.T) {
	err := createXdiamondConf()
	if err != nil {
		t.Error(err)
	}
	c := NewConfig("golang-test.1.0", SourceXdaHTTP)
	if c.Get("test2").String() != "x" {
		t.Error("配置中心http读取错误...")
	}
}

type callback struct {
}

func (c *callback) CallbackHandel(fileName string, co *ConfigObject) {
	fmt.Println("配置变更回调")
}
func TestXdiamondTCP(t *testing.T) {
	err := createXdiamondConf()
	if err != nil {
		t.Error(err)
	}
	cb := new(callback)
	SetCallbackFunc(cb)
	c := NewConfig("golang-test.1.0", SourceXdaTCP)
	if c.Get("test2").String() != "x" {
		t.Error("配置中心TCP读取错误...")
	}
}

//重置环境
func setEnv() error {
	path := os.TempDir()
	e.confDir = path + "/"
	confPath := path + "comm"
	_, err := os.Stat(confPath)
	if err != nil {
		err = os.MkdirAll(confPath, 0666)
		if err != nil {
			return err
		}
	}
	return nil
}

//设置一个测试文件
func createTestFile() error {
	fileName := e.confDir + "comm/app.toml"
	fileInfo, err := os.Create(fileName)
	defer fileInfo.Close()
	if err != nil {
		return err
	}
	confBody := `
	title = "TOML Example"
	[base]
	name = "Tom Preston-Werner"
	organization = "GitHub"
	bio = "GitHub Cofounder & CEO\nLikes tater tots and beer."
	dob = 2018-05-27T07:32:00Z # RFC3339因特网标准时间
	int = 1
	float =1.1
	bool = true
	
	[servers]
	[servers.alpha]
	ip = "10.0.0.1"
	dc = "eqdc10"

	[servers.beta]
	ip = "10.0.0.2"
	dc = "eqdc10"
	# 表嵌套
	[clients]
	data = [ ["gamma", "delta"], [1, 2] ] 

	# 数组
	hosts = [
	"alpha",
	"omega"
	]
	`
	_, err = fileInfo.WriteString(confBody)
	if err != nil {
		return err
	}
	return nil
}

// 创建用户中心配置文件
func createXdiamondConf() error {
	fileName := e.confDir + "comm/xdiamond.toml"
	fileInfo, err := os.Create(fileName)
	defer fileInfo.Close()
	if err != nil {
		return err
	}
	confBody := `
	group_id = "web"
	secret_key ="68bq57jhxmi"
	tcp_address ="10.0.200.53:5678"
	http_address ="10.0.200.53:8089"`
	_, err = fileInfo.WriteString(confBody)
	if err != nil {
		return err
	}
	return nil
}
