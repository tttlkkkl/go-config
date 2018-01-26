package conf

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	format = "json"
	uri    = "/clientapi/config"
)

type xdiamondHTTP struct {
	xdiamond
}

// 实例化配置中心http实例
func newXdiamondHTTP() *xdiamondHTTP {
	xdiamond := newXdiamond()
	return &xdiamondHTTP{xdiamond: *xdiamond}
}

// 配置中心配置解析
func (x xdiamondHTTP) analysisConfig(fileName string) (map[string]interface{}, error) {
	var err error
	var tmpSlice []interface{}
	tmpSlice, err = x.synConfigData(x.getObjectAndVersion(fileName))
	if err != nil {
		return nil, err
	}
	return x.extractKv(tmpSlice), nil
}

// 同步配置中心数据
func (x *xdiamondHTTP) synConfigData(object string, version string) ([]interface{}, error) {
	var tmp interface{}
	var err error
	var data []byte
	data, err = x.httpPull(object, version)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &tmp)
	if err != nil {
		return nil, err
	}
	tmpSlice, ok := tmp.([]interface{})
	if !ok {
		val, ok := x.getValue("success", tmp)
		if ok {
			b, ok := val.(bool)
			if ok && !b {
				return nil, errors.New("http请求错误:" + fmt.Sprintf("%v", tmp))
			}
		}
		return nil, errors.New("配置中心:类型断言失败,请检查json结构" + fmt.Sprintf("%v", tmp))
	}
	return tmpSlice, nil
}

// httpPull 从配置中心拉取数据
func (x *xdiamondHTTP) httpPull(object string, version string) ([]byte, error) {
	var response *http.Response
	var err error
	var body []byte
	response, err = http.Get(x.getFullURL(object, version))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// 获取请求地址
func (x *xdiamondHTTP) getFullURL(object string, version string) string {
	if version == "" {
		version = x.version
	}
	url := "http://" + x.HTTPAddress + uri
	url += "?groupId=" + x.GroupID + "&artifactId=" + object + "&version=" + version + "&profile=" + x.profile + "&secretKey=" + x.SecretKey + "&format=" + format
	return url
}
