package conf

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
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
	files, err := getFilesInDir(dir)
	if err != nil {
		return nil, err
	}
	var result []string
	for _, v := range files {
		ext := filepath.Ext(v.Name())
		if ext != ".toml" {
			continue
		}
		result = append(result, dir+"/"+v.Name())
	}
	return result, nil
}

//获取文件下的目录信息
func getFilesInDir(dir string) ([]os.FileInfo, error) {
	dir = strings.Replace(dir, "\\", "/", -1)
	dirInfo, err := os.OpenFile(dir, os.O_RDONLY, 0666)
	defer dirInfo.Close()
	if err != nil {
		return nil, errors.New("打开目录失败:" + dir + err.Error())
	}
	var files []os.FileInfo
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
	return files, nil
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

//tcp连接处理开始

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
	//消息协议版本
	version uint16 = 1
	//版本描述占长
	headerVersionLen = 2
	//消息长度占长
	headerLengthLen = 4
	//消息类型描述占长
	headerTypeLen = 2
	//消息头占长
	headerLen = headerVersionLen + headerLengthLen + headerTypeLen
)

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
	//客户端重载信令 0终止客户端,1重载客户端
	stopClientChanl chan int
	//心跳计时，如果间隔时间内没有收到心跳回包，尝试重新载入连接
	heartTimmer *time.Timer
}

//TCPClient tcp客户端
func tcpClient() {
	conn, err := net.Dial("tcp", e.Xdiamond.Address)
	if err != nil {
		Log.Fatal("配置中心连接失败:", err)
	}
	//defer conn.Close()
	client := &client{
		conn:              conn,
		resvResponseChanl: make(chan []byte, 100),
		resvOnewayChanl:   make(chan []byte, 100),
		stopClientChanl:   make(chan int, 3),
		heartTimmer:       time.NewTimer(clientheartInterval),
	}
	//处理连接
	go client.handelConn()
	//接收服务端数据
	go client.receivePackets()
	//心跳检测
	go client.heartCheck()
	//首次获取配置
	client.getConfig()
}

//重载客户端
func (c *client) initClient(sign int) {
	if sign != 0 && sign != 1 {
		return
	}
	if sign == 0 {
		Log.Info("关闭连接...")
		c.stopClientChanl <- 1
		c.stopClientChanl <- 1
		c.stopClientChanl <- 1
		c.heartTimmer.Stop()
		_ = c.conn.Close()
	}
	conn, err := net.Dial("tcp", e.Xdiamond.Address)
	if err != nil {
		Log.Error("连接重载失败")
	}
	_ = c.conn.Close()
	Log.Info("重载连接...")
	c.conn = conn
	//重置计时器
	c.heartTimmer.Reset(clientheartInterval)
	c.getConfig()
}

//计时时间到仍然没有心跳信令回包，前提收到心跳信令回包时要重置计时器
func (c *client) heartCheck() {
	for {
		select {
		case <-c.heartTimmer.C:
			Log.Debug("心跳超时重载...")
			c.initClient(1)
		case <-c.stopClientChanl:
			Log.Debug("退出计时器...")
			return
		}
	}
}

//从服务器接收数据并解包
func (c *client) receivePackets() {
	for {
		select {
		case sign := <-c.stopClientChanl:
			//传递信号量
			Log.Debug("退出消息协程...", sign)
			return
		default:
			unPacket(c)
		}
	}
}

//处理连接
func (c *client) handelConn() {
	for {
		select {
		//收到Oneway消息
		case data := <-c.resvOnewayChanl:
			c.handelOnewayMessage(data)
		//收到Response消息
		case data := <-c.resvResponseChanl:
			c.handelResponseMessage(data)
		case sign := <-c.stopClientChanl:
			Log.Debug("退出处理协程...", sign)
			return
		//空闲时发送心跳数据
		case <-time.Tick(heartInterval):
			c.sendHeartPacket()
		}
	}
}

//集中处理服务器返回消息
func (c *client) handelOnewayMessage(data []byte) {
	res := new(oneway)
	err := json.Unmarshal(data, res)
	Log.Debug("Response:", res)
	if err != nil {
		Log.Error("服务响应json数据解码失败:", err)
	}
	if res.Type == ONEWAY && res.Command == CONFIGCHANGED {
		Log.Info("配置有变更,准备同步配置数据...")
		c.getConfig()
	}
}

//进一步处理response类型的消息
func (c *client) handelResponseMessage(data []byte) {
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
		c.heartTimmer.Reset(clientheartInterval)
	case GETCONFIG:
		Log.Info("收到配置数据,准备更新...")
		config, ok := res.Result["configs"]
		if !ok {
			Log.Error("返回结构错误:", config)
		}
		Log.Info("更新配置数据...")
		analysisXdiamondConf(config)
	default:
		Log.Error("未知的响应类型", res.Command, "消息体:", res)
	}
}

//获取配置
func (c *client) getConfig() {
	Log.Info("更新配置....")
	c.sendDataPacket(newRequest(REQUEST, GETCONFIG))
}

//发送心跳包
func (c *client) sendHeartPacket() {
	Log.Debug("发送心跳数据包....")
	r := newRequest(REQUEST, HEARTBEAT)
	r.Data = make(auth)
	c.sendDataPacket(r)
	//发送后重置计时器
	c.heartTimmer.Reset(clientheartInterval)
}

//发送数据包
func (c *client) sendDataPacket(r *request) {
	Log.Debug("准备发送数据包:", *r)
	data, err := json.Marshal(r)
	if err != nil {
		Log.Error("消息结构序列化失败", err)
	}
	_, err = c.conn.Write(packet(r.Type, data))
	if err != nil {
		Log.Error("消息发送失败", err)
	}
}

//实例化一个请求
func newRequest(msgType messageType, cmdType commandType) *request {
	var a = auth{
		"groupId":    e.Xdiamond.GroupID,
		"artifactId": e.Xdiamond.ArtifactID,
		"version":    e.Xdiamond.Version,
		"profile":    e.Xdiamond.Profile,
		"secretKey":  e.Xdiamond.SecretKey,
	}
	r := &request{
		Type:    msgType,
		Command: cmdType,
		Data:    a,
	}
	return r
}

//解包
func unPacket(c *client) {
	header := make([]byte, headerLen)
	_, err := c.conn.Read(header)
	if err != nil {
		if err == io.EOF {
			Log.Error("远程主机主动关闭连接:", err)
		}
		// if strings.Contains(err.Error(), "use of closed network connection") {
		// }
		Log.Error("包头读取失败:", err)
		return
	}
	vs, err := getUint16(header[:2])
	if err != nil {
		Log.Error(err)
	}
	if vs != version {
		Log.Error("不支持的通信协议版本:", vs)
	}
	length, err := getUint32(header[2:6])
	if err != nil {
		Log.Error(err)
	}
	if length == 0 {
		Log.Error("错误的消息长度:", length)
	}
	msgType, err := getUint16(header[6:])
	if err != nil {
		Log.Error(err)
	}
	//读取消息体
	dataLen := length - headerTypeLen
	data := make([]byte, dataLen)
	_, err = c.conn.Read(data)
	if err != nil {
		Log.Error("消息体读取失败:", err)
	}
	if messageType(msgType) == ONEWAY {
		c.resvOnewayChanl <- data
	}
	if messageType(msgType) == RESPONSE {
		c.resvResponseChanl <- data
	}
}

//封包
func packet(msgType messageType, data []byte) []byte {
	len := len(data)
	msg := make([]byte, 0)
	msg = append(msg, pactUnit16(version)...)
	msg = append(msg, pactUint32(uint32(len)+uint32(headerTypeLen))...)
	msg = append(msg, pactUnit16(uint16(msgType))...)
	return append(msg, data...)
}

//获取2个字节的Uint16数据
func getUint16(buff []byte) (uint16, error) {
	//这里是不是2需要进一步考证
	if len(buff) != 2 {
		return 0, errors.New("无法解码的字节序")
	}
	var v uint16
	binaryBuff := bytes.NewBuffer(buff)
	//大端序
	err := binary.Read(binaryBuff, binary.BigEndian, &v)
	if err != nil {
		return 0, errors.New("Uint16解码失败:" + err.Error())
	}
	return v, nil
}

//将Unit16生成字节序
func pactUnit16(v uint16) []byte {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, v)
	return b
}

//获取4字节的uint32数据
func getUint32(buff []byte) (uint32, error) {
	//这里是不是2需要进一步考证
	if len(buff) != 4 {
		return 0, errors.New("无法解码的字节序")
	}
	var v uint32
	binaryBuff := bytes.NewBuffer(buff)
	//大端序
	err := binary.Read(binaryBuff, binary.BigEndian, &v)
	if err != nil {
		return 0, errors.New("uint32解码失败:" + err.Error())
	}
	return v, nil
}

//将Unit32生成字节序
func pactUint32(v uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, v)
	return b
}
