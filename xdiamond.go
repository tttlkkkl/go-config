package conf

import "strings"

//xdiamond xdiamond配置中心连接定义
type xdiamond struct {
	//groupId 对应groupId
	GroupID string `toml:"group_id"`
	//artifactId 对应project
	artifactID string
	//string 对应环境 取基础配置env
	profile string
	//version 配置版本,指project的版本
	version string
	//secretKey auth认证key
	SecretKey string `toml:"secret_key"`
	//address 配置中心服务器地址
	Address string
	//配置中心连接方式 支持 http,tcp
	ConnMode string `toml:"conn_mode"`
}

type xdiamondSyn interface {
	synConfigData(object string, version string) ([]interface{}, error)
}

//从接口类型的map里面获取指定数据
func (x xdiamond) getValue(key string, m interface{}) (interface{}, bool) {
	val, ok := m.(map[string]interface{})
	if ok {
		v, ok := val[key]
		if ok {
			return v, true
		}
	}
	return nil, false
}
func (x xdiamond) getObjectAndVersion(fileName string) (objectName string, version string) {
	if fileName == "" {
		return "", ""
	}
	tmp := strings.Split(fileName, ".")
	if len(tmp) == 2 {
		objectName = tmp[0]
		version = tmp[1]
	}
}
