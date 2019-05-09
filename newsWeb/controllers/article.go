package controllers

import (
	"github.com/astaxie/beego"
	"path"
	"time"
	"github.com/astaxie/beego/orm"
	"newsWeb/models"
	"math"
	_"github.com/gomodule/redigo/redis"
)

type ArticleController struct {
	beego.Controller
}

//展示首页
func(this*ArticleController)ShowIndex(){
	//获取所有文章数据，展示到页面
	o := orm.NewOrm()
	qs := o.QueryTable("Article")
	var articles []models.Article
	//qs.All(&articles)

	//获取总记录数
	count,_ := qs.Count()
	//获取总页数
	pageIndex := 2


	pageCount := math.Ceil(float64(count) / float64(pageIndex))
	//获取首页和末页数据
	//获取页码
	pageNum ,err := this.GetInt("pageNum")
	if err != nil {
		pageNum = 1
	}
	beego.Info("数据总页数未:",pageNum)

	//获取对应页的数据   获取几条数据     起始位置
	qs.Limit(pageIndex,pageIndex * (pageNum - 1)).All(&articles)






	this.Data["articles"] = articles
	this.Data["count"] = count
	this.Data["pageCount"] = pageCount
	this.Data["pageNum"] = pageNum
	this.TplName = "index.html"
}

//展示添加文章页面
func(this*ArticleController)ShowAddArticle(){
	this.TplName = "add.html"
}

//处理添加文章业务
func(this*ArticleController)HandleAddArticle(){
	//获取数据
	articleName := this.GetString("articleName")
	content := this.GetString("content")

	//校验数据
	if articleName == "" || content == "" {
		beego.Error("获取数据错误")
		this.Data["errmsg"] = "获取数据错误"
		this.TplName = "add.html"
		return
	}

	//获取图片
	//返回值 文件二进制流  文件头    错误信息
	file,head,err := this.GetFile("uploadname")
	if err != nil {
		beego.Error("获取数据错误")
		this.Data["errmsg"] = "图片上传失败"
		this.TplName = "add.html"
		return
	}
	defer file.Close()
	//校验文件大小
	if head.Size >5000000{
		beego.Error("获取数据错误")
		this.Data["errmsg"] = "图片数据过大"
		this.TplName = "add.html"
		return
	}

	//校验格式 获取文件后缀
	ext := path.Ext(head.Filename)
	if ext != ".jpg" && ext != ".png" && ext != ".jpeg" {
		beego.Error("获取数据错误")
		this.Data["errmsg"] = "上传文件格式错误"
		this.TplName = "add.html"
		return
	}

	//防止重名
	fileName := time.Now().Format("200601021504052222")


	//jianhuangcaozuo

	//把上传的文件存储到项目文件夹
	this.SaveToFile("uploadname","./static/img/"+fileName+ext)

	//处理数据
	//把数据存储到数据库
	//获取orm对象
	o := orm.NewOrm()
	//获取插入独享
	var article models.Article
	//给插入对象赋值
	article.Title = articleName
	article.Content = content
	article.Img = "/static/img/"+fileName+ext
	//插入数据
	_,err = o.Insert(&article)
	if err != nil {
		beego.Error("获取数据错误",err)
		this.Data["errmsg"] = "数据插入失败"
		this.TplName = "add.html"
		return
	}

	//返回数据  跳转页面
	this.Redirect("/index",302)
}
func (this *ArticleController)ShowContent(){
	id,err :=this.GetInt("id")
	if err!=nil{
		beego.Error("数据ID传输错误")
		this.Redirect("/index",302)
		return
	}
	o:=orm.NewOrm()
	var article models.Article
	article.Id=id
	o.Read(&article)
	article.ReadCount+=1
	o.Update(&article)
	this.Data["article"]=article
	this.TplName="content.html"
}