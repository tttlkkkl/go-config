package conf

import (
	"encoding/json"
	"io"
	"net"
	"strings"
	"sync"
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
	//重连尝试次数
	retryConnCount = 20
	//尝试重连间隔
	retryConnInterval = 5 * time.Second
)

// xdiamondTCP 配置中心tcp同步
type xdiamondTCP struct {
	xdiamond
	client
	// 暂存 "文件"信息
	fileName string
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
	conn            net.Conn
	confChangeChanl chan []interface{}
	// 客户端重载信令 0终止客户端,1重载客户端
	stopClientChanl chan int
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
	x.start()
	x.fileName = fileName
	//阻塞等待返回
	data := <-x.confChangeChanl
	//启动go携程消费无缓冲通道
	go x.synConfigData()
	return x.extractKv(data), nil
}

// 同步配置信息
func (x *xdiamondTCP) synConfigData() {
	for {
		select {
		case data := <-x.confChangeChanl:
			_ = c.genConfigObject(x.fileName, SourceXdaTCP, x.extractKv(data))
		}
	}
}

// 实例化tcp客户端
func newClient(addr string) *client {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		Log.Fatal("配置中心连接失败:", err)
	}
	//defer conn.Close()
	return &client{
		conn:            conn,
		confChangeChanl: make(chan []interface{}),
		stopClientChanl: make(chan int),
		heartTimmer:     time.NewTimer(clientheartInterval),
	}
}

// 启动客户端
func (x *xdiamondTCP) start() {
	//处理连接
	go x.handelConn()
	//心跳检测
	go x.heartCheck()
	//首次获取配置
	x.getConfig()
}

// 重载连接
func (x *xdiamondTCP) reload() {
	// 暂时停止处理协程
	x.stopClientChanl <- 1
	x.stopClientChanl <- 1

	_ = x.conn.Close()
	var wg sync.WaitGroup
	wg.Add(retryConnCount)
	t := time.NewTimer(retryConnInterval)
	stop := make(chan int, 1)
	go func() {
		for {
			select {
			case <-t.C:
				Log.Info("尝试重载连接...")
				conn, err := net.Dial("tcp", x.TCPAddress)
				wg.Done()
				if err != nil {
					Log.Error("连接重载失败...")
					continue
				}
				x.conn = conn
			case <-stop:
				wg.Add(0)
				return
			}
		}
	}()
	wg.Wait()
	Log.Info("重载连接成功...")
	//重置计时器
	x.heartTimmer.Reset(clientheartInterval)
	x.start()
}

//处理连接
func (x *xdiamondTCP) handelConn() {
	for {
		select {
		case sign := <-x.stopClientChanl:
			Log.Debug("退出处理协程...", sign)
			return
		//空闲时发送心跳数据
		case <-time.Tick(heartInterval):
			x.sendHeartPacket()
		default:
			data, msgType, err := unPacket(x.conn)
			if err != nil {
				if err == io.EOF {
					Log.Error("远程主机主动关闭连接:", err)
				}
				if strings.Contains(err.Error(), "use of closed network connection") {
					Log.Error("连接已被关闭:", err)
				}
			}
			//收到Oneway消息
			if msgType == ONEWAY {
				x.handelOnewayMessage(data)
			}
			//收到Response消息
			if msgType == RESPONSE {
				x.handelResponseMessage(data)
			}
		}
	}
}

//计时时间到仍然没有心跳信令回包，前提收到心跳信令回包时要重置计时器
func (x *xdiamondTCP) heartCheck() {
	for {
		select {
		case <-x.heartTimmer.C:
			Log.Debug("心跳超时重载...")
			x.reload()
		case <-x.stopClientChanl:
			Log.Debug("退出计时器...")
			return
		}
	}
}

//集中处理服务器返回消息
func (x *xdiamondTCP) handelOnewayMessage(data []byte) {
	res := new(oneway)
	err := json.Unmarshal(data, res)
	Log.Debug("Response:", res)
	if err != nil {
		Log.Error("服务响应json数据解码失败:", err)
	}
	if res.Type == ONEWAY && res.Command == CONFIGCHANGED {
		Log.Info("配置有变更,准备同步配置数据...")
		x.getConfig()
	}
}

//进一步处理response类型的消息
func (x *xdiamondTCP) handelResponseMessage(data []byte) {
	res := new(response)
	err := json.Unmarshal(data, res)
	Log.Debug("Response:", res)
	if err != nil {
		Log.Error("服务响应json数据解码失败:", err)
	}
	if !res.Success {
		Log.Error("服务器响应错误:", res.Error)
	}
	switch res.Command {
	case HEARTBEAT:
		Log.Debug("收到心跳回包...")
		//重载心跳检测计时器
		x.heartTimmer.Reset(clientheartInterval)
	case GETCONFIG:
		Log.Info("收到配置数据,准备更新...")
		config, ok := res.Result["configs"]
		if !ok {
			Log.Error("返回结构错误:", config)
		}
		Log.Info("更新配置数据...")
		x.confChangeChanl <- config
	default:
		Log.Error("未知的响应类型", res.Command, "消息体:", res)
	}
}

//发送心跳包
func (x *xdiamondTCP) sendHeartPacket() {
	Log.Debug("发送心跳数据包....")
	r := x.newRequest(REQUEST, HEARTBEAT)
	r.Data = make(auth)
	x.sendDataPacket(r)
	//发送后重置计时器
	x.heartTimmer.Reset(clientheartInterval)
}

//发送数据包
func (x *xdiamondTCP) sendDataPacket(r *request) {
	Log.Debug("准备发送数据包:", *r)
	data, err := json.Marshal(r)
	if err != nil {
		Log.Error("消息结构序列化失败", err)
	}
	_, err = x.conn.Write(packet(r.Type, data))
	if err != nil {
		Log.Error("消息发送失败", err)
	}
}

//获取配置
func (x *xdiamondTCP) getConfig() {
	Log.Info("更新配置....")
	x.sendDataPacket(x.newRequest(REQUEST, GETCONFIG))
}

//实例化一个请求
func (x *xdiamondTCP) newRequest(msgType messageType, cmdType commandType) *request {
	object, version := x.getObjectAndVersion(x.fileName)
	var a = auth{
		"groupId":    x.GroupID,
		"artifactId": object,
		"version":    version,
		"profile":    x.profile,
		"secretKey":  x.SecretKey,
	}
	r := &request{
		Type:    msgType,
		Command: cmdType,
		Data:    a,
	}
	return r
}
