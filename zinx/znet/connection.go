package znet

import (
	"fmt"
	"io"
	"net"
	"zinx/ziface"
   "zinx/config"
)

//具体的TCP链接模块
type Connection struct {
	Conn     *net.TCPConn
	ConnID   uint32
	isClosed bool
	//handleAPI ziface.HandleFunc
	Router ziface.IRouter
}

//初始化链接方法
func NewConnection(conn *net.TCPConn, connID uint32, router ziface.IRouter) ziface.IConnection {
	c := &Connection{
		Conn:   conn,
		ConnID: connID,
		//handleAPI: callback_api,
		Router:   router,
		isClosed: false,
	}
	return c

}

//建立针对链接读业务的方法
func (c *Connection) StartRead() {
	fmt.Println("read go is startin...")
	defer fmt.Println("connID..", c.ConnID, "Reader is exit, remote addr is = ", c.GetRemoteAddr().String())
	defer c.Stop()
	for {
		buf := make([]byte,config.GlobalObject.MaxPackageSize)
		cnt, err := c.Conn.Read(buf)
		if err != nil && err != io.EOF {
			fmt.Println("read error..", err)
			continue
		} else if cnt == 0 {
			fmt.Println("read stop...")
			break
		}
		req := NewRequest(c, buf, cnt)
		go func() {
			c.Router.Handle(req)
			c.Router.PostHandle(req)
			c.Router.PreHandle(req)
		}()
		/*err = c.handleAPI(req)
		if err != nil {
			fmt.Println("connID", c.ConnID, "handele is err", err)
			break
		}*/
	}
}
func (c *Connection) Start() {
	fmt.Println("Conn Start（）  ... id = ", c.ConnID)
	go c.StartRead()
}
func (c *Connection) Stop() {
	fmt.Println("connId stop=", c.ConnID)
	if c.isClosed == true {
		return
	}
	c.isClosed = true
	_ = c.Conn.Close()
}
func (c *Connection) GetConnID() uint32 {
	return c.ConnID
}
func (c *Connection) GetTCPConnection() *net.TCPConn {
	return c.Conn
}
func (c *Connection) GetRemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}
func (c *Connection) Send(data []byte, cnt int) error {
	_, err := c.Conn.Write(data[:cnt])
	if err != nil {
		fmt.Println("send buf error")
		return err
	}
	return nil
}
