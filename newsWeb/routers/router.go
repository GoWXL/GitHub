package routers

import (
	"newsWeb/controllers"
	"github.com/astaxie/beego"

)

func init() {
    beego.Router("/", &controllers.MainController{})
    beego.Router("/register",&controllers.UserController{},"get:ShowRegister;post:HandleRegister")
    //登录业务处理
    beego.Router("/login",&controllers.UserController{},"get:ShowLogin;post:HandleLogin")
    //首页展示
    beego.Router("/index",&controllers.ArticleController{},"get:ShowIndex")
    //添加文章业务
    beego.Router("/addArticle",&controllers.ArticleController{},"get:ShowAddArticle;post:HandleAddArticle")
    beego.Router("/content",&controllers.ArticleController{},"get:ShowContent")
}
