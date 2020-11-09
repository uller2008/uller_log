package uller

import (
	"github.com/gohouse/gorose/v2"
	"os"
	"runtime"
	"sync"
)

type LogServer struct {
	RunMod						string						// 程序运行模式  dev：开发模式，test：测试模式，prod：生产模式
	Consumer					string						// 调用者程序所在机器路径+名称
	LocalIp						string						// 本地服务器IP
	LocalPort					string						// 本地服务器监听端口
	Log							*Log						// 日志
}

func LogServerInstance() *LogServer{
	logServer := LogServer{}
	config := ConfigInstance()
	logServer.RunMod = config.IniReader.Section("run").Key("runMode").String()
	if logServer.RunMod == "" {
		panic("ini file not exist runMod")
	}
	iniLogConfig, err := config.IniReader.GetSection(logServer.RunMod + "Log")
	if err != nil {
		panic(err)
	}
	_,file,_,_ := runtime.Caller(1)
	logServer.Consumer = file
	storageEngine := StorageEngine{}
	log := Log{}
	log.StorageType = iniLogConfig.Key("storageType").String()
	if log.StorageType != "file" && log.StorageType != "db" && log.StorageType != "remotServer" {
		panic("ini file StorageType must is file,db,remotServer")
	}
	if log.StorageType == "file"{
		logFilePath := iniLogConfig.Key("logFilePath").String()
		if logFilePath == "" {
			panic("StorageType = file logFilePath can not nil")
		}
		storageEngine.File,err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil{
			panic(err)
		}
	}else if log.StorageType == "db"{
		iniLogDbConfig, err := config.IniReader.GetSection(logServer.RunMod + "LogDB")
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
	}
	sendInterval, err := iniLogConfig.Key("sendInterval").Int()
	if err != nil {
		panic("ini file not exist storageType")
	}
	log.SendInterval = sendInterval
	log.Cache, err = iniLogConfig.Key("localCache").Bool()
	if err != nil {
		panic("ini file not exist storageType")
	}
	sendCount, err := iniLogConfig.Key("sendCount").Int()
	if err != nil {
		panic("ini file not exist sendCount")
	}
	log.MaxLength = sendCount
	log.Mutex = sync.RWMutex{}
	log.StorageEngine = &storageEngine
	logServer.Log = LogInstance(log)
	whiteList := iniLogConfig.Key("whiteList").Strings(",")
	blackList := iniLogConfig.Key("blackList").Strings(",")
	if err != nil{
		panic(err)
	}
	tcpConfig := Tcp{iniLogConfig.Key("localIp").String(),iniLogConfig.Key("localPort").String(), nil}
	clientList := make(map[string]map[string]int64)
	serverTcpConfig := ServerTcp{tcpConfig,sync.RWMutex{},clientList,whiteList,blackList}
	serverTcp := ServerTcpInstance(serverTcpConfig)
	go logServer.Log.LogMonitor()
	serverTcp.Start(logServer.Log)
	return &logServer
}