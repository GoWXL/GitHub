package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/astaxie/beego/utils"
	"github.com/gomodule/redigo/redis"
	"math/rand"
	"pyg/pyg/models"
	"regexp"
	"strconv"
	"time"
)

type UserController struct {
	beego.Controller
}

//展示注册页面
func (this *UserController) ShowRegister() {
	this.TplName = "register.html"
}

//发送短信
func (this *UserController) HandleSendMsg() {
	//接收数据
	phone := this.GetString("phone")
	//定义一个容器
	resp := make(map[string]interface{})
	//返回json格式数据
	defer RespFunc(&this.Controller, resp)
	//校验数据
	if phone == "" {
		beego.Error("获取数据失败")
		resp["errno"] = 1
		resp["errmsg"] = "获取电话号码错误"

		return
	}
	//正则匹配
	reg, _ := regexp.Compile(`^1[3-9][0-9]{9}$`)
	result := reg.FindString(phone)
	if result == "" {
		resp["errno"] = 1
		resp["errmsg"] = "电话号码格式错误"
		return
	}
	//SDK调用
	//初始化客户端  需要accessKey  需要开通申请
	client, err := sdk.NewClientWithAccessKey("default", "LTAIvFXQmq69AXbm", "z4bbiK2XAx8HunIBBXWPY3JXjBI3A0")
	if err != nil {
		resp["errno"] = 2
		resp["errmsg"] = "阿里云客户端初始化失败"
		fmt.Println("阿里云客户端初始化失败")
		return
	}
	//获取6位数随机码
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	vcode := fmt.Sprintf("%06d", rnd.Int31n(1000000))

	//初始化请求对象
	request := requests.NewCommonRequest()
	request.Method = "POST"                                         //设置请求方法
	request.Scheme = "https"                                        // https | http   //设置请求协议
	request.Domain = "dysmsapi.aliyuncs.com"                        //域名
	request.Version = "2017-05-25"                                  //版本号
	request.ApiName = "SendSms"                                     //api名称
	request.QueryParams["PhoneNumbers"] = phone                     //需要发送的电话号码
	request.QueryParams["SignName"] = "真吃货之家"                       //签名名称   需要申请
	request.QueryParams["TemplateCode"] = "SMS_165115755"           //模板号   需要申请
	request.QueryParams["TemplateParam"] = `{"code":` + vcode + `}` //发送短信验证码

	response, err := client.ProcessCommonRequest(request) //发送短信
	if err != nil {
		beego.Error(err)
		resp["errno"] = 3
		resp["errmsg"] = "发送短信失败"
		return
	}
	//fmt.Println(string(response.GetHttpContentBytes()))
	//fmt.Println("结束")
	var msg Message

	json.Unmarshal(response.GetHttpContentBytes(), &msg) //解析发送结果
	if msg.Message != "OK" {
		beego.Error(err)
		resp["errno"] = 4
		resp["errmsg"] = "短信发送失败"
		return
	}
	fmt.Println(msg)

	resp["errno"] = 5
	resp["errmsg"] = "短信发送成功"
	resp["verCode"] = vcode
	fmt.Println(resp)
}

//给ajax传数据
func RespFunc(this *beego.Controller, resp map[string]interface{}) {
	this.Data["json"] = resp
	//指定传递方式
	this.ServeJSON()
}

type Message struct {
	Message   string `json:"Message"`
	RequestId string `json:"RequestId"`
	BizId     string `json:"BizId"`
	Code      string `json:"Code"`
}

//处理注册页面
func (this *UserController) HandleRegister() {
	//获取数据
	phone := this.GetString("phone")
	pwd := this.GetString("password")
	rpwd := this.GetString("repassword")
	//校验数据
	if phone == "" || pwd == "" || rpwd == "" {
		beego.Error("数据格式错误")
		this.TplName = "register.html"
		return
	}
	//处理数据
	o := orm.NewOrm()
	var user models.User
	user.Name = phone
	user.PassWord = pwd
	user.Phone = phone
	o.Insert(&user)
	//激活页面
	//*待添加cookie
	this.Ctx.SetCookie("username", phone)
	this.Redirect("/register-email", 302)
}

//展示激活页面
func (this *UserController) ShowEmailRegitser() {
	this.TplName = "register-email.html"
}

//发送邮箱激活
func (this *UserController) SendEmailRegister() {

	//获取数据
	email := this.GetString("email")
	//校验数据
	//判断空
	if email == "" {
		this.Data["errmsg"] = "电子邮箱不能为空"
		this.TplName = "register-email.html"
		return
	}

	//判断两次密码是否一致
	//校验邮箱格式
	reg, err := regexp.Compile(`^\w[\w\.-]*@[0-9a-z][0-9a-z-]*(\.[a-z]+)*\.[a-z]{2,6}$`)
	if err != nil {
		beego.Error("邮箱匹配错误", err)
		this.TplName = "register-email.html"
		return
	}
	resultEmail := reg.FindString(email)
	if resultEmail == "" {
		beego.Error("邮箱格式错误")
		this.TplName = "register-email.html"
		return
	}
	//从cookie中得到username
	username := this.Ctx.GetCookie("username")
	var user models.User
	user.Name = username
	o := orm.NewOrm()
	err = o.Read(&user, "Name")
	if err != nil {
		beego.Error("获取用户信息失败")
		this.TplName = "register-email.html"
		return
	}

	//utils 全局通用接口 工具类
	//发送激活邮件部分
	//发送邮件
	config := `{"username":"1510271838@qq.com","password":"ynojniemjvbnigch","host":"smtp.qq.com","port":587}`
	temail := utils.NewEMail(config)
	temail.To = []string{email}
	temail.From = "1510271838@qq.com"
	temail.Subject = "真吃货用户激活"

	temail.HTML = "复制该连接到浏览器中激活：127.0.0.1:8080/active?userName=" + user.Name
	//发送email
	err = temail.Send()
	if err != nil {
		beego.Error("发送邮件失败!目的邮箱是：", email, err)
		this.TplName = "register-email.html"
		return
	}
	beego.Info("发送邮件成功")
	this.Ctx.WriteString("注册成功，请前往邮箱激活!")
}

//激活邮箱
func (this *UserController) ActiveEmail() {

	//获取数据
	userName := this.GetString("userName")
	//校验数据
	if userName == "" {
		beego.Error("用户名错误")
		this.Redirect("/register-email", 302)
		return
	}

	o := orm.NewOrm()
	var user models.User
	user.Name = userName
	err := o.Read(&user, "Name")
	if err != nil {
		beego.Error("查询不到该数据")
		this.Redirect("/register-email", 302)
		return
	}
	//更新激活字段
	user.Active = true
	o.Update(&user, "Active")
	this.Redirect("/login", 302)
}

//展示用户登录页面
func (this *UserController) ShowLogin() {

	this.TplName = "login.html"
}

//用户登录
func (this *UserController) HandleLogin() {
	username := this.GetString("username")
	password := this.GetString("password")
	if password == "" || username == "" {
		beego.Error("账号密码不能为空")
		this.TplName = "login.html"
		return
	}
	var user models.User
	user.Name = username
	user.PassWord = password
	o := orm.NewOrm()
	err := o.Read(&user, "Name", "PassWord")
	beego.Info(user)
	if err != nil {
		beego.Error("没有该账号或者密码")
		this.TplName = "login.html"
		return
	}

	this.SetSession("LoginUserName", user.Name)
	this.Redirect("/index", 302)
}

//退出登录
func (this *UserController) LoginOut() {
	this.DelSession("LoginUserName")
	this.Redirect("/login", 302)
}

//展示用户中心
func (this *UserController) ShowUserCenterInfo() {
	//获取最近浏览
	conn, err := redis.Dial("tcp", "192.168.11.150:6379")
	if err != nil {
		beego.Error(err)
	}
	defer conn.Close()
	//获得5个最近浏览
	r, err := conn.Do("lrange", "recent", "0", "4")
	recent, err := redis.Strings(r, err)
	if err != nil {
		beego.Error(err)
	}

	//获取goodssku信息
	o := orm.NewOrm()
	var goods []models.GoodsSKU
	for _, v := range recent {
		var good models.GoodsSKU
		good.Id, _ = strconv.Atoi(v)
		o.Read(&good)
		goods = append(goods, good)
	}

	this.Data["goods"] = goods
	//name := this.GetSession("LoginUserName")
	this.Data["change"] = 1
	this.Data["username"] = this.GetSession("LoginUserName").(string)
	this.Layout = "user_center_layout.html"
	this.TplName = "user_center_info.html"
}

//展示收货地址页面
func (this *UserController) ShowSite() {
	//获取数据
	username := this.GetSession("LoginUserName")
	o := orm.NewOrm()
	var addr models.Address
	qs := o.QueryTable("Address").RelatedSel("User").Filter("User__Name", username.(string))
	err := qs.Filter("IsDefault", true).One(&addr)
	//校验数据
	if err != nil {
		//返回数据
		this.Data["address"] = nil
	} else {
		//返回数据
		phone := addr.Phone
		pf := phone[:3]
		pb := phone[7:11]
		addr.Phone = pf + "****" + pb
		this.Data["address"] = addr
	}

	this.Data["change"] = 3
	this.Data["username"] = this.GetSession("LoginUserName").(string)
	this.Layout = "user_center_layout.html"
	this.TplName = "user_center_site.html"
}

//添加收货地址
func (this *UserController) HandleSite() {
	//获得数据
	receiver := this.GetString("receiver")
	address := this.GetString("address")
	postCode := this.GetString("postCode")
	phone := this.GetString("phone")
	//校验数据
	if receiver == "" || address == "" || postCode == "" || phone == "" {
		this.Data["errmsg"] = "数据不能空"
		this.Redirect("/user/showsite", 302)
		return
	}
	//处理数据
	o := orm.NewOrm()
	//获取用户记录
	username := this.GetSession("LoginUserName")
	var user models.User
	user.Name = username.(string)
	err := o.Read(&user, "Name")
	if err != nil {
		beego.Error("添加收货地址，读取用户记录失败", err)
		this.Redirect("/user/showsite", 302)
		return
	}
	var addr models.Address
	addr.Receiver = receiver
	addr.Addr = address
	addr.PostCode = postCode
	addr.Phone = phone
	addr.IsDefault = true
	addr.User = &user
	//将用户旧默认地址置false
	//orm多对一查询
	var oldAddr models.Address
	qs := o.QueryTable("Address").RelatedSel("User").Filter("User__Name", username)
	err = qs.Filter("IsDefault", true).One(&oldAddr)
	//如果有默认地址
	if err == nil {
		oldAddr.IsDefault = false
		o.Update(&oldAddr, "IsDefault")
	}
	//提交数据
	//添加新的默认地址
	o.Insert(&addr)
	this.Redirect("/user/showsite", 302)
}

//展示订单页面
func (this *UserController) ShowOrders() {
	this.Data["change"] = 2
	this.Data["username"] = this.GetSession("LoginUserName").(string)
	this.Layout = "user_center_layout.html"
	this.TplName = "user_center_order.html"
}

//展示用户中心订单页
func (this *UserController) ShowUserOrder() {
	//从数据库中获取当前用户所有订单信息
	name := this.GetSession("LoginUserName")
	//获取订单信息
	o := orm.NewOrm()
	var orderinfos []models.OrderInfo
	o.QueryTable("OrderInfo").RelatedSel("User").Filter("User__Name", name.(string)).OrderBy("-Time").All(&orderinfos)

	var orders []map[string]interface{}
	for _, v := range orderinfos {
		temp := make(map[string]interface{})
		//获取当前订单所有的订单商品
		var orderGoods []models.OrderGoods
		o.QueryTable("OrderGoods").RelatedSel("OrderInfo", "GoodsSKU").Filter("OrderInfo__Id", v.Id).All(&orderGoods)

		temp["orderGoods"] = orderGoods
		temp["orderInfo"] = v

		orders = append(orders, temp)
	}

	this.Data["orders"] = orders
	this.Layout = "user_center_layout.html"
	this.TplName = "user_center_order.html"
}
