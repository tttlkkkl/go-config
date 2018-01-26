package conf

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
)

const (
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

//解包
func unPacket(conn net.Conn) (data []byte, msgType messageType, err error) {
	header := make([]byte, headerLen)
	_, err = conn.Read(header)
	if err != nil {
		// if err == io.EOF {
		// 	Log.Error("远程主机主动关闭连接:", err)
		// }
		// // if strings.Contains(err.Error(), "use of closed network connection") {
		// // }
		// Log.Error("包头读取失败:", err)
		return nil, 0, err
	}
	vs, err := getUint16(header[:2])
	if err != nil {
		return nil, 0, err
	}
	if vs != version {
		return nil, 0, errors.New("不支持的通信协议版本:" + fmt.Sprintf("%d", vs))
	}
	length, err := getUint32(header[2:6])
	if err != nil {
		return nil, 0, err
	}
	if length == 0 {
		return nil, 0, errors.New("错误的消息长度:" + fmt.Sprintf("%d", length))
	}
	t, err := getUint16(header[6:])
	if err != nil {
		return nil, 0, err
	}
	//读取消息体
	dataLen := length - headerTypeLen
	data = make([]byte, dataLen)
	_, err = conn.Read(data)
	if err != nil {
		return nil, 0, errors.New("消息体读取失败:" + fmt.Sprintf("%d", err))
	}
	return data, messageType(t), nil
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
