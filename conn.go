package conf

import (
	"io/ioutil"
	"net/http"
	"net/url"
)

//httpConn 从配置中心拉取数据
func httpConn(version string) ([]byte, error) {
	formData := make(url.Values)
	formData.Set("groupId", x.groupID)
	formData.Set("artifactId", x.artifactID)
	formData.Set("version", version)
	formData.Set("profile", x.profile)
	formData.Set("secretKey", x.secretKey)
	var response *http.Response
	var err error
	var body []byte
	response, err = http.PostForm(x.address+uri, formData)
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
