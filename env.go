package conf

import (
	"errors"
	"os"
	"runtime"
	"strings"
)

//全局环境
var e *env

const (
	// envConf 定义配置环境比如 dev test product
	envConf       = "WEB_GO_CONFIG_ENV"
	envConfigPath = "WEB_GO_CONFIG_PATH"
)

// 环境定义
type env struct {
	//配置文件路径
	confDir string
	//配置环境
	env string
}

// newEnv 初始化基本环境信息
func newEnv() (*env, error) {
	e := new(env)
	e.env = getEnv()
	var err error
	e.confDir, err = getConfigDir()
	if err != nil {
		return nil, err
	}
	return e, nil
}

// getEnv 获取配置环境
func getEnv() string {
	env := os.Getenv(envConf)
	if env == "" {
		env = "dev"
	}
	return env
}

// getConfigDir 获取配置路径
func getConfigDir() (string, error) {
	confDir := os.Getenv(envConfigPath)
	if confDir == "" {
		if runtime.GOOS == "windows" {
			confDir = "C:/web_go_config"
		} else {
			confDir = "/var/web_go_config"
		}
	}
	confDir = strings.Replace(confDir, "\\", "/", -1)
	dirInfo, err := os.Stat(confDir)
	if confDir[len(confDir)-1:] != "/" {
		confDir = confDir + "/"
	}
	if err != nil {
		return "", err
	}
	if !dirInfo.IsDir() {
		return "", errors.New(confDir + ":不是一个有效的目录")
	}
	return confDir, nil
}
