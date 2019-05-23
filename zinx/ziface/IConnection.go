package ziface

import (
	"net"
)

type IConnection interface {
	Start()
	Stop()
	//获取链接ID
	GetConnID() uint32
	//获取从哪conn原生套接字
	GetTCPConnection() *net.TCPConn
	//获取远程客户IP端地址
	GetRemoteAddr() net.Addr
	//发送数据给客户端
	Send(data []byte,cnt int) error
}

//定义业务处理接口
type HandleFunc func(request IRequest) error
