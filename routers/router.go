/*
 * @Author: yuanbaocheng
 * @Date: 2020-04-05 16:58:10
 * @LastEditors: ybc
 * @LastEditTime: 2020-09-16 17:59:08
 * @Description: file content
 */
package routers

import (
	"inspect/controllers"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/plugins/cors"
)

func init() {
	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type", "x-token"},
		ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type", "x-token"},
		AllowCredentials: true,
	}))

	ns := beego.NewNamespace("/api/v1",
		beego.NSRouter("/projects", &controllers.StatisticController{}, "get:Projects"),
		beego.NSRouter("/project", &controllers.StatisticController{}, "post:AddProject"),
		beego.NSRouter("/project", &controllers.StatisticController{}, "get:ProjectDetail"),
		beego.NSRouter("/project/test", &controllers.StatisticController{}, "post:ProjectTest"),
		beego.NSRouter("/project/statistic", &controllers.StatisticController{}, "get:ProjectChart"),
		beego.NSRouter("/project/errorLogs", &controllers.StatisticController{}, "get:ErrorLogs"),
	)
	beego.AddNamespace(ns)
	beego.SetStaticPath("/swagger", "swagger")

	// var FilterUser = func(ctx *context.Context) {
	// 	_, ok := ctx.Input.Session("uid").(int)
	// 	if !ok && ctx.Request.RequestURI != "/login" {
	// 		ctx.Redirect(302, "/login")
	// 	}
	// }

	// beego.InsertFilter("/*", beego.BeforeRouter, FilterUser)
}
