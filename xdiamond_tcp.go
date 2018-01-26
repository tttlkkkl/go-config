package conf

import (
	"net"
	"time"
)

//command 指令类型
type commandType int64

//message 消息类型
type messageType uint16

const (
	//HEARTBEAT 心跳信令
	HEARTBEAT commandType = 101
	//GETCONFIG 获取配置指令
	GETCONFIG commandType = 102
	//CONFIGCHANGED 更新配置指令，配置变更时消息类型为ONEWAY,指令类型为CONFIGCHANGED，客户端需要推送类型为CONFIGCHANGED的命令
	CONFIGCHANGED commandType = 201
)
const (
	//REQUEST 请求类型
	REQUEST messageType = iota + 1
	//RESPONSE 响应类型
	RESPONSE
	//ONEWAY 服务端通知客户端配置更新，客户端无需响应，可直接拉取配置
	ONEWAY
)
const (
	//心跳间隔,配置中心心跳间隔为15秒，为确保网络延时等特殊情况下不超时，此处设为14秒
	heartInterval = 14 * time.Second
	//客户端收取心跳回包的间隔
	clientheartInterval = 2 * heartInterval
)

// xdiamondTCP 配置中心tcp同步
type xdiamondTCP struct {
	xdiamond
	client
}

//请求体
type request struct {
	Type    messageType
	Command commandType
	Data    auth
}

//响应体
type response struct {
	Type    messageType
	Command commandType
	Success bool
	Result  map[string][]interface{}
	Error   map[string]string
}

//Oneway响应体 此消息说明配置有更新
type oneway struct {
	Type    messageType
	Command commandType
	Data    map[string]interface{}
}

//认证数据
type auth map[string]string

//客户端
type client struct {
	conn              net.Conn
	resvResponseChanl chan []byte
	resvOnewayChanl   chan []byte
	confChangeChanl   chan []interface{}
	// 客户端重载信令 0终止客户端,1重载客户端
	stopClientChanl chan int
	//客户端停止标志，1表示已停止
	stopFlag int64
	// 心跳计时，如果间隔时间内没有收到心跳回包，尝试重新载入连接
	heartTimmer *time.Timer
}

// 实例化配置中心TCP客户端
func newXdiamondTCP() *xdiamondTCP {
	xdiamond := newXdiamond()
	TCPClient := newClient(xdiamond.TCPAddress)
	return &xdiamondTCP{xdiamond: *xdiamond, client: *TCPClient}
}

// 获取并解析用户中心配置信息
func (x *xdiamondTCP) analysisConfig(fileName string) (map[string]interface{}, error) {
	x.client.start()
	//阻塞等待返回
	data := <-x.confChangeChanl
	return x.extractKv(data), nil
}

// 实例化tcp客户端
func newClient(addr string) *client {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		Log.Fatal("配置中心连接失败:", err)
	}
	//defer conn.Close()
	return &client{
		conn:              conn,
		resvResponseChanl: make(chan []byte, 100),
		resvOnewayChanl:   make(chan []byte, 100),
		confChangeChanl:   make(chan []interface{}),
		stopClientChanl:   make(chan int),
		stopFlag:          0,
		heartTimmer:       time.NewTimer(clientheartInterval),
	}
}

// 启动客户端
func (c *client) start() {

}
