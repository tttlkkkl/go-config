package conf

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
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
	//CONFIGCHANGED 更新配置指令
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
	//心跳间隔
	heartInterval = 30 * time.Second
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

//认证数据
type auth map[string]string

//客户端
type client struct {
	conn              net.Conn
	sendChanl         chan []byte
	resvChanl         chan []byte
	suspendHeartChanl chan int
	stopChanl         chan int
}

//TCPClient tcp客户端
func TCPClient() {

	fmt.Println(REQUEST, RESPONSE, ONEWAY)
	conn, err := net.Dial("tcp", e.Xdiamond.Address)
	defer conn.Close()
	if err != nil {
		fmt.Println(err)
	}
	client := &client{
		conn:              conn,
		sendChanl:         make(chan []byte, 100),
		resvChanl:         make(chan []byte, 100),
		suspendHeartChanl: make(chan int, 1),
		stopChanl:         make(chan int, 1),
	}
	go client.handelConn()
	client.receivePackets()
	client.getConfig()
}

//处理连接
func (c *client) handelConn() {
	for {
		select {
		//发送数据包
		case data := <-c.sendChanl:
			fmt.Println("发送消息", data)
		//收到完整数据包
		case data := <-c.resvChanl:
			fmt.Println("收到完整数据包", data)
		//发送心跳数据
		case <-time.Tick(heartInterval):
			c.sendHeartPacket()
		//暂停心跳数据发送
		case <-c.suspendHeartChanl:
		case <-c.stopChanl:
		}
	}

}

//获取配置
func (c *client) getConfig() {
	c.sendDataPacket(newRequest(REQUEST, GETCONFIG))
}

//发送数据包
func (c *client) sendDataPacket(r *request) {
	data, err := json.Marshal(r)
	if err != nil {
		Log.Error("请求结构序列化失败", err)
	}
	_, err = c.conn.Write(packet(r.Type, data))
	if err != nil {
		Log.Error("消息发送失败", err)
	}
}

//发送心跳包
func (c *client) sendHeartPacket() {
	c.sendDataPacket(newRequest(REQUEST, HEARTBEAT))
}

//从服务器接收数据
func (c *client) receivePackets() {
	for {
		readPacket(c.conn, c.resvChanl)
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

//通信协议封装
type protocol struct {
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

//从数据流解包
func readPacket(conn net.Conn, resvChanl chan []byte) {
	header := make([]byte, headerLen)
	_, err := conn.Read(header)
	if err != nil {
		Log.Error("包头读取失败:", err)
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
	if length != uint32(headerLengthLen) {
		Log.Error("错误的消息长度:", length)
	}
	msgType, err := getUint16(header[6:])
	if err != nil {
		Log.Error(err)
	}
	if msgType != uint16(headerTypeLen) {
		Log.Error("错误的消息类型:", msgType)
	}
	//读取消息体
	dataLen := length - uint32(msgType)
	data := make([]byte, dataLen)
	_, err = conn.Read(data)
	if err != nil {
		Log.Error("消息体读取失败:", err)
	}
	resvChanl <- data
}

//获取2个字节的Uint16数据
func getUint16(buff []byte) (uint16, error) {
	//这里是不是2需要进一步考证
	if len(buff) != 2 {
		return 0, errors.New("无法解码的字节序")
	}
	var v uint16
	binaryBuff := bytes.NewBuffer(buff)
	//小端序
	err := binary.Read(binaryBuff, binary.LittleEndian, &v)
	if err != nil {
		return 0, errors.New("Uint16解码失败:" + err.Error())
	}
	return v, nil
}

//将Unit16生成字节序
func pactUnit16(v uint16) []byte {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, v)
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
	//小端序
	err := binary.Read(binaryBuff, binary.LittleEndian, &v)
	if err != nil {
		return 0, errors.New("uint32解码失败:" + err.Error())
	}
	return v, nil
}

//将Unit32生成字节序
func pactUint32(v uint32) []byte {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint32(b, v)
	return b
}
