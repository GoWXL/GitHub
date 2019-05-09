package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"newsWeb/models"
)

type UserController struct {
	beego.Controller
}

//展示注册页面
func(this*UserController)ShowRegister(){
	this.TplName = "register.html"
}

//处理注册业务
func(this*UserController)HandleRegister(){
	//1.获取数据
	userName := this.GetString("userName")
	pwd := this.GetString("password")
	//校验数据
	if userName == "" || pwd == "" {
		beego.Error("传输数据不完整")
		this.TplName = "register.html"
		return
	}
	//处理数据
	o := orm.NewOrm()
	var user models.User
	user.Name = userName
	user.Pwd = pwd
	id,err := o.Insert(&user)
	if err != nil {
		beego.Error("用户注册失败")
		this.TplName = "register.html"
		return
	}
	beego.Info(id)
	//返回数据
	this.Redirect("/login",302)
}

//展示登录页面
func(this*UserController)ShowLogin(){
	this.TplName = "login.html"
}

//处理登录业务
func(this*UserController)HandleLogin(){
	//获取数据
	userName := this.GetString("userName")
	pwd := this.GetString("password")
	//校验数据
	if userName == "" || pwd == "" {
		beego.Error("传输数据不完整")
		this.TplName = "login.html"
		return
	}
	//处理数据
	o := orm.NewOrm()
	var user models.User
	user.Name = userName
	err := o.Read(&user,"Name")
	if err != nil {
		beego.Error("用户名不存在")
		this.TplName = "login.html"
		return
	}

	if user.Pwd != pwd{
		beego.Error("密码错误")
		this.TplName = "login.html"
		return
	}

	//返回数据
	this.Redirect("/index",302)
}