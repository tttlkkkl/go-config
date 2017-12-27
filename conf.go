package conf

const (
	format = "json"
	uri    = "/clientapi/config"
)

var x *configCenter

//configCenter xdiamond配置中心连接定义
type configCenter struct {
	//groupId 对应groupId
	groupID string
	//artifactId 对应project
	artifactID string
	//string 对应环境
	profile string
	//version 配置版本,指project的版本
	version string
	//secretKey auth认证key
	secretKey string
	//address 配置中心服务器地址
	address string
}

type conf struct {
	//cachePath 配置缓存路径
	cachePath string
	//日志路径
	logPath string
	//confPath 默认配置路径
	confPath string
}
