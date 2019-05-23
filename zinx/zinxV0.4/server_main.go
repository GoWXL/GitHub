/**
 zinx v0.1 应用
*/
package main

import (
	"fmt"
	"zinx/znet"
	"zinx/ziface"

)

//PreHandle方法  ---  用户可以在处理业务之前  自定义一些业务， 实现这个方法
//Handler方法  ---- 用户可以定义一个 业务处理的 核心方法
//PostHandle方法  --- 用户可以在处理业务之后 定义一些业务，实现这个方法
type PingRouter struct {
	znet.BaseRouter
}


func (this *PingRouter) Handle(reqeust ziface.IRequest) {
	fmt.Println("Call Router Handler...")
	//给客户端回写一个 数据
	_, err := reqeust.GetConnection().GetTCPConnection().Write([]byte("ping..ping..ping...\n"))
	if err != nil {
		fmt.Println("call  ping error")
	}
}

func (this *PingRouter) PostHandle(reqeust ziface.IRequest) {
	fmt.Println("Call Router PostHandler...")
	//给客户端回写一个 数据
	_, err := reqeust.GetConnection().GetTCPConnection().Write([]byte("after ping...\n"))
	if err != nil {
		fmt.Println("call back after ping error")
	}
}


func main() {
	//创建一个zinx server对象
	s := znet.NewServer("zinx v0.1")

	//注册一些自定义的业务
	s.AddRouter(&PingRouter{}) //真正处理核心业务的

	//让server对象 启动服务
	s.Server()

	return
}
