package main

import (
	_ "pyg/pyg/routers"
	_ "pyg/pyg/models"
	"github.com/astaxie/beego"
)

func main() {
	beego.Run()
}

