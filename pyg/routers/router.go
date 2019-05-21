package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"pyg/pyg/controllers"
)

func init() {
	beego.InsertFilter("/user/*", beego.BeforeExec, userFilterFunc)
	beego.Router("/register", &controllers.UserController{}, "get:ShowRegister;post:HandleRegister")
	beego.Router("/sendMsg", &controllers.UserController{}, "post:HandleSendMsg")
	beego.Router("/register-email", &controllers.UserController{}, "get:ShowEmailRegitser;post:SendEmailRegister")
	beego.Router("/active", &controllers.UserController{}, "get:ActiveEmail")
	beego.Router("/login", &controllers.UserController{}, "get:ShowLogin;post:HandleLogin")
	beego.Router("/index", &controllers.GoodsController{}, "get:ShowIndex")
	beego.Router("/user/loginout", &controllers.UserController{}, "get:LoginOut")
	beego.Router("/user/usercenterinfo", &controllers.UserController{}, "get:ShowUserCenterInfo")
	beego.Router("/user/showsite", &controllers.UserController{}, "get:ShowSite;post:HandleSite")
	//beego.Router("/user/showorders", &controllers.UserController{}, "get:ShowOrders")

	//生鲜首页
	beego.Router("/index_sx", &controllers.GoodsController{}, "get:ShowIndexSx")
	//商品详情
	beego.Router("/goodsDetail", &controllers.GoodsController{}, "get:ShowDetail")
	//同一类型所有商品
	beego.Router("/goodsType", &controllers.GoodsController{}, "get:ShowList")
	//查找
	beego.Router("/search", &controllers.GoodsController{}, "post:Search")
	//添加购物车
	beego.Router("/addCart", &controllers.CartController{}, "post:HandleAddCart")
	//展示购物车
	beego.Router("/user/showCart", &controllers.CartController{}, "get:ShowCart")
	//更改购物车数量-添加 减少
	beego.Router("/changeCart", &controllers.CartController{}, "post:HandleChangeCart")
	//更改购物车数量-减少
	//beego.Router("/downCart",&controllers.CartController{},"post:HandleDownCart")
	//删除购物车商品
	beego.Router("/deleteCart", &controllers.CartController{}, "post:HandleDeleteCart")
	//添加商品到订单
	beego.Router("/user/addOrder", &controllers.OrderController{}, "post:ShowOrder")
	//提交订单
	beego.Router("/pushOrder", &controllers.OrderController{}, "post:HandlePushOrder")
	//展示用户中心订单页
	beego.Router("/user/userOrder", &controllers.UserController{}, "get:ShowUserOrder")
	//支付
	beego.Router("/pay", &controllers.OrderController{}, "get:Pay")
	//支付完成
	beego.Router("/payOK", &controllers.OrderController{}, "get:PayOK;post:PayOKNotify")
}

func userFilterFunc(ctx *context.Context) {
	name := ctx.Input.Session("LoginUserName")
	if name == nil {
		beego.Error("过滤器：用户没有登录")
		ctx.Redirect(302, "/index")
	}
}
