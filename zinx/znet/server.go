package znet

import (
	"fmt"
	"net"
	"zinx/config"
	"zinx/ziface"
)

type Server struct {
	//服务器IP
	IPVersion string
	IP        string
	//服务器PORT
	Port int
	//服务器名字
	Name   string
	Router ziface.IRouter
}

//定义一个回显业务
func CallBackBusi(request ziface.IRequest) error  {
	fmt.Println("【conn Handle】 CallBack..")
	c := request.GetConnection().GetTCPConnection()
	buf := request.GetData()
	cnt := request.GetDataLen()
	_, err := c.Write(buf[:cnt])
	if err != nil {
		fmt.Println("write back err", err)
		return err
	}
	return nil
}

//初始化new方法
func NewServer(name string) ziface.IServer {
	s := &Server{
		Name:      config.GlobalObject.Name,
		IPVersion: "tcp4",
		IP:        config.GlobalObject.Host,
		Port:      config.GlobalObject.Port,
		Router:    nil,
	}
	return s
}
func (s *Server) Start() {
	fmt.Printf("[start] server Listenner at IP :%s,Port :%d,is starting..\n", s.IP, s.Port)
	addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
	if err != nil {
		fmt.Println("resolve tcp addr error:", err)
		return
	}
	listenner, err := net.ListenTCP(s.IPVersion, addr)
	if err != nil {
		fmt.Println("listn", s.IPVersion, "err", err)
		return
	}
	var cid uint32
	cid = 0
	go func() {
		for {
			conn, err := listenner.AcceptTCP()
			if err != nil {
				fmt.Println("Accepterr", err)
				continue
			}
			//创建一个connertion的对象
			dealConn := NewConnection(conn, cid, s.Router)
			cid++
			go dealConn.Start()

		}
	}()

}
func (s *Server) Stop() {

}
func (s *Server) Server() {
	s.Start()
	select {}
}
func (s *Server) AddRouter(router ziface.IRouter) {
	s.Router = router
}
