/*
 * @Author: yuanbaocheng
 * @Date: 2020-04-11 17:35:02
 * @LastEditors: ybc
 * @LastEditTime: 2020-09-10 16:45:55
 * @Description: file content
 */

package controllers

import (
	"encoding/json"
	"github.com/asaskevich/govalidator"
	"github.com/astaxie/beego"
)

type ResponseData struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type BaseController struct {
	beego.Controller
}

func (this *BaseController) ReturnResult(data interface{}) {
	this.Data["json"] = &ResponseData{
		Code: 0,
		Msg:  "ok",
		Data: data,
	}
	this.ServeJSON()
}

func (this *BaseController) ReturnError(code int, msg string) {
	this.Data["json"] = &ResponseData{
		Code: code,
		Msg:  msg,
		Data: "",
	}
	this.ServeJSON()
}

func (this *BaseController) ReturnErrorMsg(msg string) {
	this.ReturnError(-1, msg)
	return
}

func (this *BaseController) ParseJsonValid(param interface{}) error {

	if err := json.Unmarshal(this.Ctx.Input.RequestBody, &param); err != nil {
		this.ReturnErrorMsg("parseJson:" + err.Error())
		return err
	}

	_, err := govalidator.ValidateStruct(param)
	if err != nil {
		this.ReturnErrorMsg(err.Error())
		return err
	}
	return nil
}

func (this *BaseController) ParseFormValid(param interface{}) error {
	if err := this.ParseForm(param); err != nil {
		this.ReturnErrorMsg("parseForm:" + err.Error())
		return err
	}

	_, err := govalidator.ValidateStruct(param)
	if err != nil {
		this.ReturnErrorMsg(err.Error())
		return err
	}
	return nil
}
