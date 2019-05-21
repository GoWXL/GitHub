package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/gomodule/redigo/redis"
	"github.com/smartwalle/alipay"
	"pyg/pyg/models"
	"strconv"
	"strings"
	"time"
)

type OrderController struct {
	beego.Controller
}

func (this *OrderController) ShowOrder() {
	//获取数据
	goodsIds := this.GetStrings("checkGoods")
	//校验数据
	if len(goodsIds) == 0 {
		this.Redirect("/user/showCart", 302)
		return
	}
	//处理数据
	//获取当前用户的所有收货地址
	name := this.GetSession("LoginUserName")

	o := orm.NewOrm()
	var addrs []models.Address
	o.QueryTable("Address").RelatedSel("User").Filter("User__Name", name.(string)).All(&addrs)
	this.Data["addrs"] = addrs

	conn, err := redis.Dial("tcp", "192.168.11.150:6379")
	if err != nil {
		beego.Error(err)
		return
	}

	//获取商品,获取总价和总件数
	var goods []map[string]interface{}
	var totalPrice, totalCount int

	for _, v := range goodsIds {
		temp := make(map[string]interface{})
		id, _ := strconv.Atoi(v)
		var goodsSku models.GoodsSKU
		goodsSku.Id = id
		o.Read(&goodsSku)

		//获取商品数量
		count, _ := redis.Int(conn.Do("hget", "cart_"+name.(string), id))

		//计算小计
		littlePrice := count * goodsSku.Price

		//把商品信息放到行容器
		temp["goodsSku"] = goodsSku
		temp["count"] = count
		temp["littlePrice"] = littlePrice

		totalPrice += littlePrice
		totalCount += 1

		goods = append(goods, temp)

	}

	//返回数据
	this.Data["totalPrice"] = totalPrice
	this.Data["totalCount"] = totalCount
	this.Data["truePrice"] = totalPrice + 10
	this.Data["goods"] = goods
	this.Data["goodsIds"] = goodsIds
	this.TplName = "place_order.html"
}

//提交订单
func (this *OrderController) HandlePushOrder() {
	//获取数据
	addrId, err1 := this.GetInt("addrId")
	payId, err2 := this.GetInt("payId")
	goodsIds := this.GetString("goodsIds")
	totalCount, err3 := this.GetInt("totalCount")
	totalPrice, err4 := this.GetInt("totalPrice")

	resp := make(map[string]interface{})
	defer RespFunc(&this.Controller, resp)

	name := this.GetSession("LoginUserName")
	if name == nil {
		resp["errno"] = 2
		resp["errmsg"] = "当前用户未登录"
		return
	}

	//校验数据
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || goodsIds == "" {
		resp["errno"] = 1
		resp["errmsg"] = "传输数据不完整"
		return
	}
	//处理数据
	//把数据插入到mysql数据库中
	//获取用户对象和地址对象
	o := orm.NewOrm()
	var user models.User
	user.Name = name.(string)
	o.Read(&user, "Name")

	var address models.Address
	address.Id = addrId
	o.Read(&address)

	var orderInfo models.OrderInfo

	orderInfo.User = &user
	orderInfo.Address = &address
	orderInfo.PayMethod = payId
	orderInfo.TotalCount = totalCount
	orderInfo.TotalPrice = totalPrice
	orderInfo.TransitPrice = 10
	orderInfo.OrderId = time.Now().Format("20060102150405" + strconv.Itoa(user.Id))
	//开启事务
	o.Begin()
	o.Insert(&orderInfo)

	conn, _ := redis.Dial("tcp", "192.168.11.150:6379")

	defer conn.Close()
	//插入订单商品
	//goodsIds  //2  3  5
	goodsSlice := strings.Split(goodsIds[1:len(goodsIds)-1], " ")
	for _, v := range goodsSlice {
		//插入订单商品表

		//获取商品信息
		id, _ := strconv.Atoi(v)
		var goodsSku models.GoodsSKU
		goodsSku.Id = id
		o.Read(&goodsSku)

		oldStock := goodsSku.Stock
		beego.Info("原始库存等于", oldStock)

		//获取商品数量
		count, _ := redis.Int(conn.Do("hget", "cart_"+name.(string), id))

		//获取小计
		littlePrice := goodsSku.Price * count

		//插入
		var orderGoods models.OrderGoods
		orderGoods.OrderInfo = &orderInfo
		orderGoods.GoodsSKU = &goodsSku
		orderGoods.Count = count
		orderGoods.Price = littlePrice
		//插入之前需要更新商品库存和销量
		if goodsSku.Stock < count {
			resp["errno"] = 4
			resp["errmsg"] = "库存不足"
			o.Rollback()
			return
		}
		//goodsSku.Stock -= count
		//goodsSku.Sales += count
		time.Sleep(time.Second * 5) //手动加延迟

		o.Read(&goodsSku)

		qs := o.QueryTable("GoodsSKU").Filter("Id", id).Filter("Stock", oldStock)
		num, _ := qs.Update(orm.Params{"Stock": goodsSku.Stock - count, "Sales": goodsSku.Sales + count})
		if num == 0 {
			resp["errno"] = 7
			resp["errmsg"] = "购买失败，请重新排队！"
			o.Rollback()
			return
		}

		_, err := o.Insert(&orderGoods)
		if err != nil {
			resp["errno"] = 3
			resp["errmsg"] = "服务器异常"
			o.Rollback()
			return
		}
		_, err = conn.Do("hdel", "cart_"+name.(string), id)
		if err != nil {
			resp["errno"] = 6
			resp["errmsg"] = "清空购物车失败"
			o.Rollback()
			return
		}
	}

	//返回数据
	o.Commit()
	resp["errno"] = 5
	resp["errmsg"] = "OK"
}

//支付
func (this *OrderController) Pay() {
	//获取数据
	orderId, err := this.GetInt("orderId")
	if err != nil {
		this.Redirect("/user/userOrder", 302)
		return
	}
	//处理数据
	o := orm.NewOrm()
	var orderInfo models.OrderInfo
	orderInfo.Id = orderId
	o.Read(&orderInfo)

	//支付

	//appId, aliPublicKey, privateKey string, isProduction bool
	publiKey := `MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwENGy4iRgaMQSl2+p2ew
rYrpkn8EJt88WQd/Lr5hqQJui3KvvQOJu68YjBzDpD0IoO7415LbtgzGm/vBpNQM
gVHQ4IKZgdgTILimAbR6CmOcGa5zKVIk0/IG7CgCTki5HD/TJF3YNx+Vk688Ulac
1YqxL6xAwaTuCxAlI6Tcpbd+N/miQZEHXFH3MnbReFCz7T/JmFx4leV7lWCzJAXF
/n7RvbghpRA71B526ibjni8V5nWxHzinLFT0MuXoikkXfLDhLQG1MADPwnEijwpX
pDi1EqOEZaZZPxIg8JYC38/1u+EBXsTESieieK84rbt9FCp+RxyEGRpMqW9SLkvL
3wIDAQAB`

	privateKey := `MIIEowIBAAKCAQEAwENGy4iRgaMQSl2+p2ewrYrpkn8EJt88WQd/Lr5hqQJui3Kv
vQOJu68YjBzDpD0IoO7415LbtgzGm/vBpNQMgVHQ4IKZgdgTILimAbR6CmOcGa5z
KVIk0/IG7CgCTki5HD/TJF3YNx+Vk688Ulac1YqxL6xAwaTuCxAlI6Tcpbd+N/mi
QZEHXFH3MnbReFCz7T/JmFx4leV7lWCzJAXF/n7RvbghpRA71B526ibjni8V5nWx
HzinLFT0MuXoikkXfLDhLQG1MADPwnEijwpXpDi1EqOEZaZZPxIg8JYC38/1u+EB
XsTESieieK84rbt9FCp+RxyEGRpMqW9SLkvL3wIDAQABAoIBABkA5tUTZrDwTu8M
7/1/a6e2GBg4MocHoyaE5hJjKfo72bqC6L3xFtl0tQGLwBm84kFjsrL+Y1pyoOWq
QQ25kgLDbCG2elY7jolD2jsAiJqPR77DRDDMgQObRzExJtOde41j84aYOcU5c09o
i7S9lNnklpR3l1hXpamEqP/Qse1PDOM8VieC6gHQA/KwujJtgT1tgveMmiyVDFGa
4uQBh0gnfS1NEI3QLVpDHHTIn9MiYtP3eg1XeEFPg9x1Gxt3Z8sN+fDVVe4zSLH8
BTXJ9S31izwnZy9zjjr9MAkWRveTXGwrKY0LdlpLZrzDn8p6CUJA5xXrYVDyrLsH
Y9oQpjECgYEA7RN+gqZNjOT/xva4jfVJX3Mgl1Se+bUczVgRxPXX4yq7ujic2QSz
SkG4Bti5d/oJq+jtnk1jEDc98Vn2sWSlXbz/d5HzzW1mJe1xy3u8ViWUDajbYPJZ
//VWH4qBttFQlc5Wvzrdm5i21LvmfTne0j6IRZ5itQ9yJ+jM8Et5kFkCgYEAz5wM
jECr/UhfnzJSUP5WPtpV+IhqktERdjO4dpvcBz2lrSMuRWLbTTrM1xiS24ivwz9z
xLNU7QxYThbe1CFE1TfgHoU9Vv+xEtx0eeSu53SmPXQ3HNjE1Y/qthbdLcEdL5Vd
ofrnJaMAIuK1jWdfPk3sq9qaLGGDcuLP5Bws9vcCgYBkzbcrIj8zO2OuW9WZNsSd
+zvOXMLD9khq35men9HN26u6wLugYylA17TB5IDoDL70A7SVbN5EVNjXuKL2Ro8x
zlzpoHuDy5J1agLKvLAWCSBstnGhRSsTdGPMQX5qF5ImQHgOE5+Ku2JyDfsxH9wo
lUIoJ/JcflbRtWD+g3kK8QKBgQCVbeluLcJdTPFegXbUSyxCkx5cA7xJrmeWH4X/
ARHuuEV+iBru4Eeen9r+WvahQxHXQ92Mz9Mpx7/rfPSn1MZZfZ03+oj7DJEkVT8U
2S+28rQQ+YwNnEyYtrymkXBjVWMvc5/wTcp/wYIAmhM5ExVvn+DglThxB0L4tx4R
PuJyYwKBgE4WmldadNgrxtWVB2TZb3jSV9hQKaQTQ8GLME+Z5dikWI9MG5ea0X6F
Ol6bNlGcdN1pjB2REeyNbfSTKMmBm8qKEYVe7eurJFLdXoZ9hJao1P7WLgH0PZzq
3k5NxhJ/OU+RPvStRuq7L1u9nUPBPRsHIq0oC4w+8U7l83aWUiMU
`

	client := alipay.New("2016093000634101", publiKey, privateKey, false)
	var p = alipay.TradePagePay{}
	p.NotifyURL = "http://127.0.0.1:8080/payOK"
	p.ReturnURL = "http://127.0.0.1:8080/payOK"
	p.Subject = "品优购"
	p.OutTradeNo = orderInfo.OrderId
	p.TotalAmount = strconv.Itoa(orderInfo.TotalPrice)

	p.ProductCode = "FAST_INSTANT_TRADE_PAY"
	url, err := client.TradePagePay(p)
	if err != nil {
		beego.Error(err)
		beego.Error("支付失败")
		return
	}
	//更新订单状态
	//orderInfo.Orderstatus = 1

	payUrl := url.String()
	beego.Info("url是：", payUrl)
	this.Redirect(payUrl, 302)
}

//支付完成 支付宝异步通知
func (this *OrderController) PayOKNotify()  {
	//需要公网ip
	beego.Info("异步post通知！！！！！！！！！！！！！！！！！！！！！！！")
	this.Redirect("/index", 302)
}

//支付完成 用户跳转通知
func (this *OrderController) PayOK() {
	//获取orderId和交易号
	orderId := this.GetString("out_trade_no")
	traderNo := this.GetString("trade_no")
	o := orm.NewOrm()
	var order models.OrderInfo
	order.OrderId = orderId
	beego.Info("支付跳转的orderId是：", orderId)
	err := o.Read(&order, "OrderId")
	if err != nil {
		beego.Error(err)
		this.Redirect("/index", 302)
		return
	}
	order.Orderstatus = 1
	order.TradeNo = traderNo
	_, err = o.Update(&order)
	if err != nil {
		beego.Error(err)
		this.Redirect("/index", 302)
		return
	}
	this.Redirect("/index", 302)
}