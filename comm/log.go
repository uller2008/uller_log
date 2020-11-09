package uller

import (
	"encoding/json"
	"fmt"
	"github.com/gohouse/gorose/v2"
	"net"
	"os"
	"sync"
	"time"
)

// 日志队列
type Log struct {
	Consumer				string					// 调用者程序所在机器路径+名称
	StorageType				string					// 日志保存类型  file：本地文件存储，db：数据库存储，remotServer：远程服务器存储
	StorageEngine			*StorageEngine			// 日志存储引擎
	Cache					bool					// 本地是否缓存
	SendInterval			int						// 开启缓存后日志存储间隔单位秒
	LogData					[]LogData				// 日志数组
	Mutex 					sync.RWMutex
	MaxLength				int						// 日志缓存数量，开启缓存后当日志数量到达一定值后忽略SendInterval值将日志存储
	Secret					string					// 远程服务器存储传输过程加密秘钥，空值则传输不加密
	CliendId				string					// 服务器端下发客户端id
}

// 存储引擎
type StorageEngine struct {
	File					*os.File				// 文件存储
	OrmEngine				*gorose.Engin			// 数据库存储引擎
	Tcp						*net.TCPConn			// 远程服务器存储tcp长链接
}

// 远程服务器存储配置
type StorageTcp struct {
	Tcp						*net.TCPConn			// 远程服务器存储tcp长链接

}

// 日志
type LogData struct {
	Md5Key					string					`gorose:"md5_key" json:"md5_key"`			//IP、Level、Logger、Exception、LogMessage生成的md5Key值
	IP						string					`gorose:"ip" json:"ip"`						//错误产生机器IP
	Consumer				string					`gorose:"consumer" json:"consumer"`			//程序所在机器路径+名称
	Lev						string					`gorose:"lev" json:"lev"`					//错误级别
	Logger					string					`gorose:"logger" json:"logger"`				//错误产生者，可以是类名.方法名
	LogMessage				string					`gorose:"log_message" json:"log_message"`	//错误日志
	Exception				string					`gorose:"exception" json:"exception"`		//错误信息
	CreateTime				string					`gorose:"create_time" json:"create_time"`	//创建时间
	SendInterval			int						`gorose:"send_interval" json:"send_interval"`//记录间隔单位秒
	Count					int						`gorose:"count" json:"count"`				//日志数量
}

func LogInstance(logConfig ...Log) *Log{
	log := Log{}
	if len(logConfig) > 0{
		log = logConfig[0]
	}else{
		filePath := GetCurrentDirectory() + "/log.log"
		file,err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			panic(err)
		}
		storageEngine := StorageEngine{file,nil,nil}
		log = Log{"unknown","file",&storageEngine,true,60,nil,sync.RWMutex{},100,"",""}
	}
	return &log
}

func (log *Log)GetMod5(logData *LogData)(md5Key string){
	return GetMD5Encode(logData.IP + string(logData.Lev) + logData.Logger + logData.LogMessage + logData.Exception)
}

func(log *Log)Exit(logData *LogData,mod ...bool)(exist bool){
	exist = false
	if logData.Md5Key == ""{
		logData.Md5Key = log.GetMod5(logData)
	}
	for i:=0;i<len(log.LogData);i++{
		if log.LogData[i].Md5Key == logData.Md5Key {
			exist = true
			if len(mod) > 0{
				log.LogData[i].Count = log.LogData[i].Count + logData.Count
			}
			break
		}
	}
	return
}

func(log *Log)Push(logData []LogData)(length int){
	if log.Cache {
		exit := false
		length = len(log.LogData)
		log.Mutex.Lock()
		defer log.Mutex.Unlock()
		for i:=0;i<len(logData);i++{
			if logData[i].CreateTime == ""{
				logData[i].CreateTime = time.Now().Format("2006-01-02 15:04:05")
			}
			exit = log.Exit(&logData[i],true)
			if !exit {
				log.LogData = append(log.LogData,logData[i])
				length++
				if length == log.MaxLength{
					log.Write()
					length = 0
				}
			}
		}
	}else{
		log.Write()
	}
	return length
}

func (log *Log)Write()(){
	if len(log.LogData) > 0 {
		if log.StorageType == "file" {
			_, err := log.WriteFile(log.StorageEngine.File)
			if err != nil {
				fmt.Println(err)
			}
		} else if log.StorageType == "db" {
			_, err := log.WriteDb(log.StorageEngine.OrmEngine)
			if err != nil {
				fmt.Println(err)
			}
		} else if log.StorageType == "remotServer" {
			_, err := log.WriteTcpConn()
			if err != nil {
				fmt.Println(err)
			}
		}
		log.LogData = nil
	}
}

func(log *Log)WriteFile(f *os.File)(num int,err error){
	num = 0
	data,err := json.Marshal(log.LogData)
	if err != nil{
		return
	}
	num,err = f.Write(data)
	if err != nil{
		return
	}
	return
}

func(log *Log)WriteDb(db *gorose.Engin)(num int64,err error){
	orm := db.NewOrm()
	num,err = orm.Table("log").Data(log.LogData).Insert()
	if err != nil {
		fmt.Println(err)
	}
	return
}

func(log *Log)WriteTcpConn()(num int,err error){
	protocol := ProtocolInstance()
	secret := false
	if len(log.Secret) >0 {
		secret = true
	}
	msgBytes,_ := json.Marshal(log.LogData)
	if err != nil{
		return
	}
	if log.Secret != ""{
		msgBytes,err = Rc4Encrypt(log.Secret,msgBytes)
		if err != nil{
			fmt.Println(err)
		}
	}
	headerBytes,_ := protocol.HeaderPack(cLog,log.CliendId,len(msgBytes),log.Cache,secret)
	protocol.SendPack(log.StorageEngine.Tcp,BytesCombine(headerBytes,msgBytes))
	return
}

func (log *Log)LogMonitor(){
	for  {
		time.Sleep(time.Duration(log.SendInterval) * time.Second)
		log.Mutex.Lock()
		log.Write()
		log.Mutex.Unlock()
	}
}