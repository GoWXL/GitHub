package config

import (
	"encoding/json"
	"io/ioutil"
)

type GlobalObj struct {
	//当前的监听IP
	Host string
	//当前监听的端口
	Port int
	//当前zinxserver的名称
	Name string
	//当前框架的版本号
	Version string
	//每次read的最大长度
	MaxPackageSize uint32
}

//定义一个全局对外的配置对象
var GlobalObject *GlobalObj
//添加一个加载配置文件的方法
func (g *GlobalObj) LoadConfig() {
	data, err := ioutil.ReadFile("conf/zinx.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(data, &GlobalObject)
	if err != nil {
		panic(err)
	}
}
func init() {
	GlobalObject = &GlobalObj{
		Name:           "ZinxServerApp",
		Host:           "0.0.0.0",
		Port:           8999,
		Version:        "zinxV0.4",
		MaxPackageSize: 512,
	}
	//加载文件
GlobalObject.LoadConfig()

}
