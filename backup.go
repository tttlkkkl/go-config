package conf

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
)

const (
	backupsDir = "comm/___backups___/"
)

// 备份配置
func backups(fileName string, confMap map[string]interface{}) error {
	fileName, err := getFullBackFileName(fileName)
	if err != nil {
		return err
	}
	jsonData, err := json.Marshal(confMap)
	if err != nil {
		return errors.New("备份失败,无法序列化配置数据..." + err.Error())
	}
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0664)
	if err != nil {
		return errors.New("备份失败,备份文件打开失败..." + err.Error())
	}
	defer file.Close()
	_, err = file.Write(jsonData)
	if err != nil {
		return errors.New("备份失败,配置数据写入失败..." + err.Error())
	}
	return nil
}

// 备份恢复
func backupRecovery(fileName string) (map[string]interface{}, error) {
	fileName, err := getFullBackFileName(fileName)
	if err != nil {
		return nil, err
	}
	jsonData, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, errors.New("从备份文件读取失败..." + err.Error())
	}
	//Log.Fatal(string(jsonData))
	var tmp map[string]interface{}
	err = json.Unmarshal(jsonData, &tmp)
	if err != nil {
		return nil, errors.New("配置解码失败..." + err.Error())
	}
	return tmp, nil
}

// 本地备份文件全名
func getFullBackFileName(fileName string) (string, error) {
	dir := e.confDir + backupsDir
	dirInfo, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(dir, 0664)
			if err != nil {
				return "", errors.New("备份目录创建失败..." + err.Error())
			}
		} else {
			return "", err
		}
	} else if !dirInfo.IsDir() {
		return "", errors.New(dir + " : 不是一个有效的目录")
	}
	return dir + fileName + ".back", nil
}
