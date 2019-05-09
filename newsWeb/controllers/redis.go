package controllers

import (
	"github.com/astaxie/beego"
	"github.com/gomodule/redigo/redis"
	_ "github.com/gomodule/redigo/redis"
)
func init(){
	conn,err:=redis.Dial("tcp",":6379")
	if err!=nil{
		beego.Error("redis连接失败")
		return
	}
	defer conn.Close()
	//resp,err:=conn.Do("get","c1")
	//result,_:=redis.String(resp,err)
	//beego.Info("获取的数据为：",result)
	resp,err:=conn.Do("mget","t1","t2","t3")
	result,_:=redis.Values(resp,err)
	var v1,v3 int
	var v2 string
	redis.Scan(result,&v1,&v2,&v3)
	beego.Info("获取的数据：",v1,v2,v3)
}