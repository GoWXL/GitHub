package znet

import "zinx/ziface"

type Request struct {
	//链接信息
	conn ziface.IConnection
	//数据内容
	data []byte
	//数据长度
	len int
}
func NewRequest(conn ziface.IConnection,data []byte,len int)ziface.IRequest{
	req:=&Request{
		conn:conn,
		data:data,
		len:len,
	}
	return req
}
//得到当前请求的链接
func (r *Request) GetConnection() ziface.IConnection {
	return r.conn
}

//得到链接的数据
func (r *Request) GetData() []byte {
	return r.data
}

//得到数据的长度
func (r *Request) GetDataLen() int {
	return r.len
}
