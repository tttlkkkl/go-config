package conf

import (
	"errors"
	"flag"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

//httpConn 从配置中心拉取数据
func httpConn() ([]byte, error) {
	var response *http.Response
	var err error
	var body []byte
	url := "http://" + e.Xdiamond.Address + uri
	url += "?groupId=" + e.Xdiamond.GroupID + "&artifactId=" + e.Xdiamond.ArtifactID + "&version=" + e.Xdiamond.Version + "&profile=" + e.Xdiamond.Profile + "&secretKey=" + e.Xdiamond.SecretKey
	response, err = http.Get(url)
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

//获取一个目录下的所有toml配置文件路径
func getTomlFilesInDir(dir string) ([]string, error) {
	dir = strings.Replace(dir, "\\", "/", -1)
	dirInfo, err := os.OpenFile(dir, os.O_RDONLY, 0666)
	defer dirInfo.Close()
	if err != nil {
		return nil, errors.New("打开目录失败:" + dir + err.Error())
	}
	var files []os.FileInfo
	var result []string
	for {
		//每次读取50个文件名
		r, err := dirInfo.Readdir(50)
		if err != nil && err != io.EOF {
			return nil, errors.New("打开目录失败:" + dir + err.Error())
		}
		if err == io.EOF {
			break
		}
		files = append(files, r...)
	}
	for _, v := range files {
		ext := filepath.Ext(v.Name())
		if ext != ".toml" {
			continue
		}
		result = append(result, dir+"/"+v.Name())
	}
	return result, nil
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
	conf = flag.String("conf", "", "额外载入一个配置文件或配置文件目录")
	flag.Parse()
	return *env, *conf
}

//获取app运行目录
func getAppRuntimeDir() (string, error) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	return strings.Replace(dir, "\\", "/", -1), err
}
