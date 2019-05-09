package models

import (
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

type User struct {
	Id int
	Name string
	Pwd string
}

//文章和类型是什么关系
//用户和文章是什么关系  多对多  单表操作  primary key

type Article struct {
	Id int `orm:"pk;auto"`
	Title string `orm:"unique;size(40)"`
	Content string `orm:"size(500)"`
	Img string	`orm:"null"`
	Time time.Time `orm:"type(datetime);auto_now_add"`
	ReadCount int	`orm:"default(0)"`
}

func init(){
	//注册数据库
	orm.RegisterDataBase("default","mysql","root:123456@tcp(127.0.0.1:3306)/newWeb")
	//注册表
	orm.RegisterModel(new(User),new(Article))
	//跑起来
	orm.RunSyncdb("default",false,true)
}
