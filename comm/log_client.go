package uller

import (
	"encoding/json"
	"errors"
	"github.com/gohouse/gorose/v2"
	"io"
	"net"
	"os"
	"runtime"
	"sync"
	"time"
)

type LogClient struct {
	RunMod						string						// 程序运行模式  dev：开发模式，test：测试模式，prod：生产模式
	Consumer					string						// 调用者程序所在机器路径+名称
	PingInterval				int							// 客户端发送心跳包间隔，单位秒
	LocalIp						string						// 本机ip地址
	CliendId					string						// 服务器端下发客户端id
	Log							*Log						// 日志
}

func LogClientInstance() *LogClient{
	logClient := LogClient{}
	config := ConfigInstance()
	logClient.RunMod = config.IniReader.Section("run").Key("runMode").String()
	logClient.LocalIp = GetLocalIP()
	if logClient.RunMod == "" {
		panic("ini file not exist runMod")
	}
	_,file,_,_ := runtime.Caller(1)
	logClient.Consumer = file
	iniLogConfig, err := config.IniReader.GetSection(logClient.RunMod + "Log")
	if err != nil {
		panic(err)
	}
	storageEngine := StorageEngine{}
	log := Log{}
	log.StorageType = iniLogConfig.Key("storageType").String()
	if log.StorageType != "file" && log.StorageType != "db" && log.StorageType != "remotServer" {
		panic("ini file StorageType must is file,db,remotServer")
	}
	log.SendInterval, err = iniLogConfig.Key("sendInterval").Int()
	log.CliendId = logClient.CliendId
	if err != nil {
		panic("ini file not exist storageType")
	}
	log.Cache, err = iniLogConfig.Key("localCache").Bool()
	if err != nil {
		panic("ini file not exist storageType")
	}
	log.MaxLength, err = iniLogConfig.Key("sendCount").Int()
	if err != nil {
		panic("ini file not exist sendCount")
	}
	encrypSecret := iniLogConfig.Key("encrypSecret").String()
	log.Secret = encrypSecret
	log.Mutex = sync.RWMutex{}
	log.StorageEngine = &storageEngine
	logClient.Log = LogInstance(log)
	if log.StorageType == "file"{
		logFilePath := iniLogConfig.Key("logFilePath").String()
		if logFilePath == "" {
			panic("StorageType = file logFilePath can not nil")
		}
		storageEngine.File,err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil{
			panic(err)
		}
		go logClient.Log.LogMonitor()
	}else if log.StorageType == "db"{
		iniLogDbConfig, err := config.IniReader.GetSection(logClient.RunMod + "LogDB")
		if err != nil{
			panic(err)
		}
		maxOpenConns,_ := iniLogDbConfig.Key("maxOpenConns").Int()
		maxIdleConns,_ := iniLogDbConfig.Key("maxIdleConns").Int()
		var config = &gorose.Config{Driver:iniLogDbConfig.Key("drive").String(),Dsn:iniLogDbConfig.Key("dsn").String(),Prefix:"",SetMaxOpenConns:maxOpenConns,SetMaxIdleConns:maxIdleConns}
		storageEngine.OrmEngine, err = gorose.Open(config)
		if err != nil{
			panic(err)
		}
		go logClient.Log.LogMonitor()
	}else if log.StorageType == "remotServer"{
		remotServerIP := iniLogConfig.Key("remotServerIP").String()
		remotServerPort := iniLogConfig.Key("remotServerPort").String()
		logClient.PingInterval,err = iniLogConfig.Key("pingInterval").Int()
		if err != nil{
			panic(err)
		}
		if log.StorageType == "remotServer" && (remotServerIP == "" || remotServerPort == "") {
			panic("StorageType = remotServer remotServerIP and remotServerPort can not nil")
		}
		clientTcp := ClientTcp{Tcp{remotServerIP,remotServerPort,nil} }
		clientTcp.Start()
		storageEngine.Tcp = clientTcp.Tcp.TcpConn
		err = logClient.SendProfile(storageEngine.Tcp)
		if err != nil{
			panic(err)
		}
		go logClient.SendPing(storageEngine.Tcp)
	}
	return &logClient
}

func (logClient *LogClient)Debug(logger string,logMessage string,exception string)(err error){
	logData := []LogData{}
	logData = append(logData,LogData{"",logClient.LocalIp,logClient.Consumer,"debug",logger,logMessage,exception,"" ,logClient.Log.SendInterval,1})
	logClient.Log.Push(logData)
	return
}

func (logClient *LogClient)SendProfile(conn *net.TCPConn)(err error){
	protocol := Protocol{}
	clientProfile := ClientProfile{}
	clientProfile.Consumer = logClient.Consumer
	clientProfile.IP = logClient.LocalIp
	clientProfile.PingInterval = logClient.PingInterval
	clientProfile.Secret = logClient.Log.Secret
	msgBytes,_ := json.Marshal(clientProfile)
	headerBytes,_ := protocol.HeaderPack(cProfile,"",len(msgBytes),false,false)
	sendBytes := BytesCombine(headerBytes,msgBytes)
	conn.Write(sendBytes)
	timestamp := time.Now().UnixNano() / 1e6
	Loop:
	for{
		_,err := conn.Read(protocol.Header.Cmd[0:])
		if err != nil || err == io.EOF{
			continue
		}
		conn.Read(protocol.Header.ClientId[0:])
		conn.Read(protocol.Header.PackLen[0:])
		conn.Read(protocol.Header.Timestamp[0:])
		conn.Read(protocol.Header.ClientCache[0:])
		conn.Read(protocol.Header.ClientEncrypt[0:])
		if BytesToInt(protocol.Header.PackLen) > 0{
			msg := make([]byte,BytesToInt(protocol.Header.PackLen))
			conn.Read(msg)
			err =  errors.New(string(msg))
			break Loop
		}
		cmd := string(protocol.Header.Cmd[0:])
		switch {
		case cmd == sProfile :
			logClient.CliendId = string(protocol.Header.ClientId[0:])
			break Loop
		default:
			if timestamp < time.Now().UnixNano() / 1e6 - 10 * 1000{
				break Loop
			}
		}
	}
	return
}

func (logClient * LogClient)SendPing(conn *net.TCPConn){
	ping := Ping{logClient.Consumer,logClient.PingInterval}
	for{
		msgBytes,_ := json.Marshal(ping)
		protocol := ProtocolInstance()
		headerBytes,_ := protocol.HeaderPack(cPing,logClient.CliendId,len(msgBytes),false,false)
		conn.Write(BytesCombine(headerBytes,msgBytes))
		time.Sleep(time.Duration(logClient.PingInterval) * time.Second)
	}
}