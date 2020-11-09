package uller

import (
	"fmt"
	"net"
	"sync"
)

type Tcp struct {
	IP						string
	Port					string							// 端口
	TcpConn					*net.TCPConn					// tcp链接
}

type ServerTcp struct {
	Tcp						Tcp
	Mutex 					sync.RWMutex
	ClientList				map[string]map[string]int64		// 链接到服务器的客户端列表
	WhiteList				[]string						// 服务器端ip白名单
	BlackList				[]string						// 服务器端ip黑名单
}

type ClientTcp struct {
	Tcp						Tcp
}

func ServerTcpInstance(tcpConfig ...ServerTcp)(tcp *ServerTcp){
	if len(tcpConfig) > 0{
		tcp = &tcpConfig[0]
	}else{
		tcp = new(ServerTcp)
		tcp.Tcp.IP = GetLocalIP()
		tcp.Tcp.Port = "8419"
	}
	return
}

func (serverTcp * ServerTcp)Start(log *Log){
	tcpAddr,err := net.ResolveTCPAddr("tcp",serverTcp.Tcp.IP + ":" + serverTcp.Tcp.Port)
	if err != nil{
		panic(err)
		return
	}
	tcpListener,err := net.ListenTCP("tcp",tcpAddr)
	if err != nil {
		panic(err)
		return
	}
	defer tcpListener.Close()
	for{
		serverTcp.Tcp.TcpConn,err = tcpListener.AcceptTCP()
		if err!=nil {
			fmt.Println(err)
			continue
		}
		go serverTcp.ServerPipe(serverTcp.Tcp.TcpConn,log)
	}
}

func (serverTcp * ServerTcp)ServerPipe(conn *net.TCPConn,log *Log){
	defer conn.Close()
	remoteAddr := conn.RemoteAddr().String()
	if serverTcp.BlackList != nil {
		if StringArrSearch(serverTcp.BlackList,remoteAddr) >= 1{
			conn.Close()
			return
		}
	}else if serverTcp.WhiteList != nil{
		if StringArrSearch(serverTcp.BlackList,remoteAddr) == -1{
			conn.Close()
			return
		}
	}
	protocolInstance := ProtocolInstance()
	protocolInstance.UnPack(serverTcp,log)
}

func ClientTcpInstance(tcpConfig ...ClientTcp)(tcp *ClientTcp){
	if len(tcpConfig) > 0{
		tcp = &tcpConfig[0]
	}else{
		tcp = new(ClientTcp)
		tcp.Tcp.IP = GetLocalIP()
		tcp.Tcp.Port = "8419"
	}
	return
}

func (clientTcp * ClientTcp)Start(){
	tcpAddr,err := net.ResolveTCPAddr("tcp",clientTcp.Tcp.IP + ":" + clientTcp.Tcp.Port)
	clientTcp.Tcp.TcpConn,err = net.DialTCP("tcp",nil,tcpAddr)
	if err!=nil {
		panic(err)
		return
	}
}

