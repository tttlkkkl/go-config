package conf

import (
	"errors"
	"strings"

	"github.com/BurntSushi/toml"
)

type localFile struct {
}

func newLocalFile() *localFile {
	return new(localFile)
}

// 解析本地配置文件
func (l *localFile) analysisConfig(fileName string) (map[string]interface{}, error) {
	var data = make(map[string]interface{})
	var err error
	fileName, err = l.getFullFileName(fileName)
	if err != nil {
		return nil, err
	}
	_, err = toml.DecodeFile(fileName, &data)
	if err != nil {
		return nil, errors.New("配置文件" + fileName + "解析失败:" + err.Error())
	}
	return data, nil
}

// 获取配置文件全名
func (l localFile) getFullFileName(fileName string) (string, error) {
	if fileName == "" {
		return "", errors.New("未指定配置文件名称")
	}
	fileName = strings.Replace(fileName, ".", "/", -1)
	if fileName[:1] == "/" {
		fileName = fileName[1:]
	}
	fileName = e.confDir + fileName + ".toml"
	return fileName, nil
}
