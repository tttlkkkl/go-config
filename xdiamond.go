package conf

import (
	"strings"

	"github.com/BurntSushi/toml"
)

//xdiamond xdiamond配置中心连接定义
type xdiamond struct {
	//groupId 对应groupId
	GroupID string `toml:"group_id"`
	//string 对应环境 取基础配置env
	profile string
	//version 配置版本,指project的版本
	version string
	//secretKey auth认证key
	SecretKey string `toml:"secret_key"`
	//TCPAddress 配置中心TCP地址
	TCPAddress string `toml:"tcp_address"`
	//HTTPAddress 配置中心http地址
	HTTPAddress string `toml:"http_address"`
}

//初始化配置中心基本配置
func newXdiamond() *xdiamond {
	x := new(xdiamond)
	xdiamondConfFileName := e.confDir + "comm/xdiamond.toml"
	_, err := toml.DecodeFile(xdiamondConfFileName, x)
	if err != nil {
		Log.Fatal(err)
	}
	x.profile = e.env
	x.version = "1.0"
	return x
}

// 提取有效的kv
func (x *xdiamond) extractKv(s []interface{}) map[string]interface{} {
	var kvMapTmp = make(map[string]interface{})
	for _, v := range s {
		val, ok := x.getValue("config", v)
		if !ok {
			continue
		}
		key, keyOk := x.getValue("key", val)
		value, valueOk := x.getValue("value", val)
		if keyOk && valueOk {
			key, ok := key.(string)
			if !ok {
				continue
			}
			kvMapTmp[key] = value
		}
	}
	return kvMapTmp
}

// 从接口类型的map里面获取指定数据
func (x *xdiamond) getValue(key string, m interface{}) (interface{}, bool) {
	val, ok := m.(map[string]interface{})
	if ok {
		v, ok := val[key]
		if ok {
			return v, true
		}
	}
	return nil, false
}

// 获取配置中心配置对象和版本
func (x *xdiamond) getObjectAndVersion(fileName string) (objectName string, version string) {
	if fileName == "" {
		return "", ""
	}
	index := strings.Index(fileName, ".")
	if index != -1 {
		objectName = fileName[:index]
		version = fileName[index+1:]
	}
	objectName = fileName
	version = x.version
	return objectName, version
}
