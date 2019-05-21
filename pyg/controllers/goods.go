package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/gomodule/redigo/redis"
	"math"
	"pyg/pyg/models"
	"strconv"
)

type GoodsController struct {
	beego.Controller
}

//func (this *GoodsController) ShowIndex() {
//	//获取session用户名
//	name := this.GetSession("LoginUserName")
//	if name != nil {
//		this.Data["loginusername"] = name.(string)
//	} else {
//		this.Data["loginusername"] = ""
//	}
//	this.TplName = "index.html"
//}

func (this *GoodsController) ShowIndex() {
	name := this.GetSession("LoginUserName")
	if name != nil {
		this.Data["loginusername"] = name.(string)
	} else {
		this.Data["loginusername"] = ""
	}

	//获取类型信息并传递给前段
	//获取一级菜单
	o := orm.NewOrm()
	//接受对象
	var oneClass []models.TpshopCategory
	//查询
	o.QueryTable("TpshopCategory").Filter("Pid", 0).All(&oneClass)

	//获取2级分类
	var types []map[string]interface{}
	for _, v := range oneClass {
		t := make(map[string]interface{})
		var temp []models.TpshopCategory
		o.QueryTable("TpshopCategory").Filter("Pid", v.Id).All(&temp)
		//1级分类的一个对象
		t["t1"] = v
		//对应的2级分类的对象切片
		t["t2"] = temp
		types = append(types, t)
	}

	//获得3级分类
	for _, v := range types {
		var secondTypes []map[string]interface{}
		for _, v2 := range v["t2"].([]models.TpshopCategory) {
			t := make(map[string]interface{})
			var temp []models.TpshopCategory
			o.QueryTable("TpshopCategory").Filter("Pid", v2.Id).All(&temp)
			t["t21"] = v2
			t["t22"] = temp
			secondTypes = append(secondTypes, t)
		}
		//第三级，是[]map[string]interface{}
		v["t3"] = secondTypes
	}

	//给前端传值
	this.Data["types"] = types
	this.TplName = "index.html"
}

func (this *GoodsController) ShowIndexSx() {
	//获取生鲜首页内容
	o := orm.NewOrm()
	//获取所有类型
	var goodsTypes []models.GoodsType
	o.QueryTable("GoodsType").All(&goodsTypes)
	this.Data["goodsTypes"] = goodsTypes

	//获取轮播图
	var goodsBanners []models.IndexGoodsBanner
	o.QueryTable("IndexGoodsBanner").OrderBy("Index").All(&goodsBanners)
	this.Data["goodsBanners"] = goodsBanners

	//获取促销商品
	var promotionBanners []models.IndexPromotionBanner
	o.QueryTable("IndexPromotionBanner").OrderBy("Index").All(&promotionBanners)
	this.Data["promotions"] = promotionBanners

	//获取首页商品展示
	var goods []map[string]interface{}
	for _, v := range goodsTypes {
		var textGoods []models.IndexTypeGoodsBanner
		var imageGoods []models.IndexTypeGoodsBanner
		qs := o.QueryTable("IndexTypeGoodsBanner").
			RelatedSel("GoodsType", "GoodsSKU").Filter("GoodsType__Id", v.Id).OrderBy("Index")
		//获取文字商品
		qs.Filter("DisplayType", 0).All(&textGoods)
		qs.Filter("DisplayType", 1).All(&imageGoods)
		//定义行容器
		temp := make(map[string]interface{})
		temp["goodsType"] = v
		temp["textGoods"] = textGoods
		temp["imageGoods"] = imageGoods

		goods = append(goods, temp)
	}
	this.Data["goods"] = goods

	this.Layout = "sxlayout.html"
	this.TplName = "index_sx.html"
}

//商品详情页
func (this *GoodsController) ShowDetail() {
	////获取数据
	//id,err := this.GetInt("Id")
	////校验数据
	//if err != nil {
	//	beego.Error("商品链接错误")
	//	this.Redirect("/index_sx",302)
	//	return
	//}
	////处理数据
	////根据id获取商品有关数据
	//o := orm.NewOrm()
	//var goodsSku models.GoodsSKU
	//goodsSku.Id = id
	//o.Read(&goodsSku)
	//
	////获取所有类型
	//var goodsTypes []models.GoodsType
	//o.QueryTable("GoodsType").All(&goodsTypes)
	//this.Data["goodsTypes"] = goodsTypes
	//
	////获取详情
	//var goods models.Goods
	//o.QueryTable("Goods").Filter("GoodsSKU__Id", id).One(&goods)
	//this.Data["goods"] = goods
	//
	////获取同一类型的新品推荐
	//var newGoods []models.GoodsSKU
	//qs := o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Name",goods.GoodsType.Name)
	//qs.OrderBy("-Time").Limit(2,0).All(&newGoods)
	//this.Data["newGoods"] = newGoods
	//
	//beego.Info(newGoods, goodsSku.GoodsType.Name)
	//
	////传递数据
	//this.Data["goodsSku"] = goodsSku
	//this.TplName = "detail.html"

	//获取数据
	id, err := this.GetInt("Id")
	//校验数据
	if err != nil {
		beego.Error("商品链接错误")
		this.Redirect("/index_sx", 302)
		return
	}
	//处理数据
	//根据id获取商品有关数据
	o := orm.NewOrm()
	var goodsSku models.GoodsSKU
	/*goodsSku.Id = id
	o.Read(&goodsSku)*/
	//获取商品详情
	o.QueryTable("GoodsSKU").RelatedSel("Goods", "GoodsType").Filter("Id", id).One(&goodsSku)

	//获取同一类型的新品推荐
	var newGoods []models.GoodsSKU
	qs := o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Name", goodsSku.GoodsType.Name)
	qs.OrderBy("-Time").Limit(2, 0).All(&newGoods)
	this.Data["newGoods"] = newGoods

	//添加最近浏览物品到redis中存储
	conn, err := redis.Dial("tcp", "192.168.11.150:6379")
	if err != nil {
		beego.Error(err)
	}
	defer conn.Close()

	//删除旧数据
	conn.Do("lrem", "recent", "0", goodsSku.Id)
	_, err = conn.Do("lpush", "recent", goodsSku.Id)
	if err != nil {
		beego.Error(err)
	}


	//传递数据
	this.Data["goodsSku"] = goodsSku
	this.Layout = "sxlayout.html"
	this.TplName = "detail.html"
}

//展示商品列表页
func (this *GoodsController) ShowList() {
	//获取数据
	id, err := this.GetInt("id")
	//校验数据
	if err != nil {
		beego.Error("类型不存在")
		this.Redirect("/index_sx", 302)
		return
	}
	//处理数据
	o := orm.NewOrm()
	var goods []models.GoodsSKU
	//获取排序方式
	sort := this.GetString("sort")
	//实现分页
	qs := o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id", id)
	//获取总页码
	count, _ := qs.Count()
	pageSize := 1
	pageCount := int(math.Ceil(float64(count) / float64(pageSize)))
	//获取当前页码
	pageIndex, err := this.GetInt("pageIndex")
	if err != nil {
		pageIndex = 1
	}
	pages := PageEdit(pageCount, pageIndex)
	this.Data["pages"] = pages
	//获取上一页，下一页的值
	var prePage, nextPage int
	//设置个范围
	if pageIndex-1 <= 0 {
		prePage = 1
	} else {
		prePage = pageIndex - 1
	}

	if pageIndex+1 >= pageCount {
		nextPage = pageCount
	} else {
		nextPage = pageIndex + 1
	}

	this.Data["prePage"] = prePage
	this.Data["nextPage"] = nextPage

	qs = qs.Limit(pageSize, pageSize*(pageIndex-1))

	//获取排序
	if sort == "" {
		qs.All(&goods)
	} else if sort == "price" {
		qs.OrderBy("Price").All(&goods)
	} else {
		qs.OrderBy("-Sales").All(&goods)
	}

	this.Data["sort"] = sort

	//返回数据
	this.Data["pageIndex"] = strconv.Itoa(pageIndex)
	this.Data["id"] = id
	this.Data["goods"] = goods
	this.Layout = "sxlayout.html"
	this.TplName = "list.html"
}

func PageEdit(pageCount int, pageIndex int) []int {
	//不足五页
	var pages []int
	if pageCount < 5 {
		for i := 1; i <= pageCount; i++ {
			pages = append(pages, i)
		}
	} else if pageIndex <= 3 {
		for i := 1; i <= 5; i++ {
			pages = append(pages, i)
		}
	} else if pageIndex >= pageCount-2 {
		for i := pageCount - 4; i <= pageCount; i++ {
			pages = append(pages, i)
		}
	} else {
		for i := pageIndex - 2; i <= pageIndex+2; i++ {
			pages = append(pages, i)
		}
	}

	return pages
}

//查找商品
func (this *GoodsController) Search() {
	//获取数据
	wanted := this.GetString("search")
	//校验数据
	if wanted == "" {
		beego.Info("查询数据为空")
		this.Redirect("/index_sx", 302)
		return
	}
	//处理数据
	o := orm.NewOrm()
	var goods []models.GoodsSKU
	o.QueryTable("GoodsSKU").Filter("Name__contains", wanted).All(&goods)

	//返回数据
	this.Data["goods"] = goods
	this.Layout = "sxlayout.html"
	this.TplName = "search.html"
}
