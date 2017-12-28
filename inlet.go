package conf

import (
	"errors"
	"flag"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

//httpConn 从配置中心拉取数据
func httpConn(version string) ([]byte, error) {
	if version == "" {
		version = e.Xdiamond.Version
	}
	formData := make(url.Values)
	formData.Set("groupId", e.Xdiamond.GroupID)
	formData.Set("artifactId", e.Xdiamond.ArtifactID)
	formData.Set("version", version)
	formData.Set("profile", e.Xdiamond.Profile)
	formData.Set("secretKey", e.Xdiamond.SecretKey)
	var response *http.Response
	var err error
	var body []byte
	response, err = http.PostForm(e.Xdiamond.Address+uri, formData)
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

//fileConn 从文件获取配置数据
func fileConn(file string) ([]byte, error) {
	if file == "" {
		return nil, errors.New("无效的文件路径")
	}
	return ioutil.ReadFile(file)
}

//getOption 获取命令行参数
func getOption() (envFile string, defaultLoad string) {
	defaultEnvFile, err := getAppRuntimeDir()
	if err != nil {
		defaultEnvFile = ""
	} else {
		defaultEnvFile = defaultEnvFile + "/.env.toml"
	}
	var env, conf *string
	env = flag.String("env", defaultEnvFile, "环境配置文件")
	conf = flag.String("conf", "", "额外载入一个配置文件")
	flag.Parse()
	return *env, *conf
}

//获取app运行目录
func getAppRuntimeDir() (string, error) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	return strings.Replace(dir, "\\", "/", -1), err
}
