package uller

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

const (
	cLog string = "v1.0"		// 通讯包:v1.0
	cPing string = "ping"		// 心跳包:ping
	cProfile string = "ctpf"	// 链接创建身份识别包:ctpf
	sProfile string = "stpf"	// 服务器下发客户端标识:stpf
)

// 数据包通讯协议
type Protocol struct {
	Header 						Header							// 包头
	Msg            				[]byte					    	// 日志数据
}

// 包头
type Header struct{
	Cmd        					[4]byte							// 通讯命令
	ClientId					[32]byte						// 服务器下发客户端标识，可以是加盐加密串防止长链被其他程序使用
	PackLen						[4]byte							// 发送Msg包长
	Timestamp      				[8]byte   						// 13位时间戳
	ClientCache					[1]byte							// 客户端是否缓存 0x00 false 0x01 true
	ClientEncrypt				[1]byte							// 客户端是否加密 0x00 false 0x01 true
}

// 客户端链接识别包
type ClientProfile struct {
	Consumer					string							// 客户端标识(使用程序所在服务器全路径+程序名)
	IP							string							// 客户端ip
	PingInterval				int								// 客户端发送心跳包间隔，单位秒
	Secret						string							// 客户端传输加密秘钥
}

// 心跳包
type Ping struct {
	Consumer					string							// 发送者
	PingInterval				int								// 心跳间隔单位秒
}

func ProtocolInstance(protocolConfig ...Protocol) *Protocol{
	protocol := Protocol{}
	if len(protocolConfig) > 0{
		protocol = protocolConfig[0]
	}
	return &protocol
}

// 包头封包
func (p *Protocol)HeaderPack(cmd string,clientId string,packLen int,clientCache bool,clientEncrypt bool)(buf []byte ,err error){
	if len(cmd) != 4{
		err = errors.New("protocol cmd length err")
		return
	}
	if clientId == ""{
		clientId = GetMD5Encode("")
	}
	cmdBytes := []byte(cmd)
	for i:=0;i<len(p.Header.Cmd);i++{
		p.Header.Cmd[i] = cmdBytes[i]
	}
	clientIdBytes := []byte(clientId)
	for i:=0;i<len(p.Header.ClientId);i++{
		p.Header.ClientId[i] = clientIdBytes[i]
	}
	p.Header.PackLen = IntToBytes(packLen)
	p.Header.Timestamp = Int64ToBytes(time.Now().UnixNano() / 1e6)
	p.Header.ClientCache = [1]byte{0x01}
	if clientCache{
		p.Header.ClientCache = [1]byte{0x01}
	}
	p.Header.ClientEncrypt = [1]byte{0x00}
	if clientEncrypt{
		p.Header.ClientEncrypt = [1]byte{0x01}
	}
	buf = BytesCombine(p.Header.Cmd[0:],p.Header.ClientId[0:],p.Header.PackLen[0:],p.Header.Timestamp[0:],p.Header.ClientCache[0:],p.Header.ClientEncrypt[0:])
	return
}

// 解包
func (p *Protocol)UnPack(serverTcp * ServerTcp,log *Log)(){
	defer serverTcp.Tcp.TcpConn.Close()
	clientIp := serverTcp.Tcp.TcpConn.RemoteAddr().String()
	_,err := serverTcp.Tcp.TcpConn.Read(p.Header.Cmd[0:])
	if err != nil || err == io.EOF{
		return
	}
	serverTcp.Tcp.TcpConn.Read(p.Header.ClientId[0:])
	serverTcp.Tcp.TcpConn.Read(p.Header.PackLen[0:])
	serverTcp.Tcp.TcpConn.Read(p.Header.Timestamp[0:])
	serverTcp.Tcp.TcpConn.Read(p.Header.ClientCache[0:])
	serverTcp.Tcp.TcpConn.Read(p.Header.ClientEncrypt[0:])
	msg := make([]byte,BytesToInt(p.Header.PackLen))
	serverTcp.Tcp.TcpConn.Read(msg)
	clientProfile := ClientProfile{}
	err = json.Unmarshal(msg,&clientProfile)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	var res string
	clientId := GetMD5Encode(clientIp + clientProfile.Consumer)
	if _, ok := serverTcp.ClientList[clientId]; ok {
		res = clientIp + "can not finded client profile,server closed."
		resBytes := []byte(res)
		headerBytes,err := p.HeaderPack(sProfile,"",len(resBytes),false,false)
		if err != nil{
			return
		}
		serverTcp.Tcp.TcpConn.Write(BytesCombine(headerBytes,resBytes))
		return
	}else{
		headerBytes,err := p.HeaderPack(sProfile,clientId,0,false,false)
		if err != nil{
			return
		}
		newClient := make(map[string]int64)
		newClient["PingTimeStamp"] = BytesToInt64(p.Header.Timestamp)
		newClient["PingInterval"] = int64(clientProfile.PingInterval)
		serverTcp.ClientList[clientId] = newClient
		serverTcp.Tcp.TcpConn.Write(headerBytes)
	}
	Loop:
	for{
		_,err := serverTcp.Tcp.TcpConn.Read(p.Header.Cmd[0:])
		if err != nil || err == io.EOF{
			continue
		}
		serverTcp.Tcp.TcpConn.Read(p.Header.ClientId[0:])
		serverTcp.Tcp.TcpConn.Read(p.Header.PackLen[0:])
		serverTcp.Tcp.TcpConn.Read(p.Header.Timestamp[0:])
		serverTcp.Tcp.TcpConn.Read(p.Header.ClientCache[0:])
		serverTcp.Tcp.TcpConn.Read(p.Header.ClientEncrypt[0:])
		msg := make([]byte,BytesToInt(p.Header.PackLen))
		serverTcp.Tcp.TcpConn.Read(msg)
		cmd := string(p.Header.Cmd[0:])
		switch {
		case cmd == cLog :
			logData := []LogData{}
			if p.Header.ClientEncrypt[0] == 0x01 && clientProfile.Secret != ""{
				msg,err = Rc4Decrypt(clientProfile.Secret,msg)
				if err != nil{
					break Loop
				}
			}else if p.Header.ClientEncrypt[0] == 0x01 && clientProfile.Secret == ""{
				break Loop
			}
			err = json.Unmarshal(msg,&logData)
			if err != nil {
				fmt.Println(err.Error())
				break Loop
			}
			log.Push(logData)
		case cmd == cPing :
			ping := Ping{}
			err := json.Unmarshal(msg,&ping)
			if err != nil {
				fmt.Println(err.Error())
				continue
			}
			if _, ok := serverTcp.ClientList[clientId]; ok {
				serverTcp.ClientList[clientId]["PingTimeStamp"] = BytesToInt64(p.Header.Timestamp)
				serverTcp.ClientList[clientId]["PingInterval"] = int64(ping.PingInterval)
				fmt.Println(serverTcp.ClientList)
			}else{
				break Loop
			}
		default:
			// 心跳间隔监测，退出长时间没有心跳包的client
			_, hasKey := serverTcp.ClientList[clientId]
			if hasKey{
				if serverTcp.ClientList[clientId]["PingTimeStamp"] < time.Now().UnixNano() / 1e6 - 2000 * int64(clientProfile.PingInterval){
					delete(serverTcp.ClientList,clientId)
					break Loop
				}
			}else{
				break Loop
			}
		}
	}
	return
}

// 发包
func (p *Protocol)SendPack(conn *net.TCPConn,buf []byte)error{
	wr := bufio.NewWriter(conn)
	_,err := wr.Write(buf)
	if err != nil{
		return err
	}
	err = wr.Flush()
	if err != nil{
		return err
	}
	return err
}