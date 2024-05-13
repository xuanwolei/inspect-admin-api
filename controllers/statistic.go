/*
 * @Author: yuanbaocheng
 * @Date: 2020-04-11 17:14:00
 * @LastEditors: yuanbaocheng
 * @LastEditTime: 2020-10-23 15:12:05
 * @Description: file content
 */
package controllers

import (
	"encoding/json"
	"fmt"
	"inspect/models"
	"inspect/services"
	"net"
	"strings"
	"time"

	"github.com/astaxie/beego"
)

// Operations about object
type StatisticController struct {
	BaseController
}

type keep struct {
	ResponseStatus json.Number `json:"response_status"`
	Time           string      `json:"time"`
	NetworkCode    json.Number `json:"network_code"`
	KeepServerIp   string      `json:"keep_server_ip"`
}

type keepChart struct {
	Time     string      `json:"time"`
	Count    json.Number `json:"count"`
	Min      json.Number `json:"min"`
	Max      json.Number `json:"max"`
	Mean     json.Number `json:"mean"`
	Median   json.Number `json:"median"`
	ErrorNum json.Number `json:"error_num"`
}
type statusChart struct {
	Time           string      `json:"time"`
	Count          json.Number `json:"count"`
	ResponseStatus string      `json:"response_status"`
}

type projectData struct {
	Id          int    `json:"id"`
	Name        string `json:"name" valid:"required,stringlength(1|100)"`
	Hosts       string `json:"hosts" valid:"required"`
	HttpHost    string `json:"http_host"`
	Type        string `json:"type" valid:"required"`
	Paths       string `json:"paths" valid:"required,stringlength(1|3000)"`
	Level       int    `json:"level" valid:"required"`
	Phone       string `json:"phone"`
	Timeout     int    `json:"timeout" valid:"required"`
	Desc        string `json:"desc"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	NoticeToken string `json:"notice_token"`
}

type testProjectData struct {
	Id       int      `json:"id"`
	Name     string   `json:"name" valid:"required,stringlength(1|100)"`
	Hosts    []string `json:"hosts" valid:"required"`
	HttpHost string   `json:"http_host"`
	Type     string   `json:"type" valid:"required"`
	Paths    []paths  `json:"paths" valid:"required"`
	Level    int      `json:"level" valid:"required"`
	Timeout  int      `json:"timeout" valid:"required"`
	Desc     string   `json:"desc"`
}

type paths struct {
	Title        string `json:"title"`
	Method       string `json:"method"`
	Data         string `json:"data"`
	Path         string `json:"path"`
	ValidateType int    `json:"validate_type"`
	ValidateRule string `json:"validate_rule"`
	TrueHttpCode int    `json:"true_http_code"`
}

type tcpRsponse struct {
	ErrorCode    int    `json:"error_code"`
	ResponseBody string `json:"response_body"`
}

//图表查询参数
type chartParam struct {
	ProjectId     string `form:"project_id" valid:"required"`
	StartTime     string `form:"start_time"`
	OverTime      string `form:"over_time"`
	TimeGroup     string `form:"time_group" valid:"stringlength(1|5)"`
	RequestNumber string `form:"request_number"`
	RequestHost   string `form:"request_host"`
	RequestPath   string `form:"request_path"`
	KeepServerIp  string `form:"keep_server_ip"`
	Page          int    `form:"page"`
	Limit         int    `form:"limit"`
	ErrorStatus   string `form:"error_status"`
}

func (o *StatisticController) Projects() {
	page, _ := o.GetInt("page")
	name := o.GetString("name")
	var (
		where map[string]interface{} = make(map[string]interface{})
	)
	if name != "" {
		where["name"] = name
	}
	re, err := models.GetProjects(where, page)
	if err != nil {
		o.ReturnError(-1, err.Error())
		return
	}
	o.ReturnResult(re)
}

func (o *StatisticController) AddProject() {
	var param projectData
	if err := o.ParseJsonValid(&param); err != nil {
		return
	}

	if param.Id == 0 {
		res := services.MysqlInstance.Table("config").NewRecord(param)

		fmt.Println("res:", res)
		param.CreatedAt = time.Now().Unix()
		services.MysqlInstance.Table("config").Create(&param)
	} else {
		param.UpdatedAt = time.Now().Format("2006-01-02 15:04:05")
		services.MysqlInstance.Table("config").Where("id = ?", param.Id).Update(&param)
	}
	o.ReturnResult(param)
}

//测试
func (o *StatisticController) ProjectTest() {
	var param testProjectData
	if err := o.ParseJsonValid(&param); err != nil {
		return
	}
	conn, err := net.Dial("tcp", beego.AppConfig.String("inspect.server.addr"))
	if err != nil {
		o.ReturnError(-1, err.Error())
		return
	}
	defer conn.Close()

	send := map[string]interface{}{
		"event": "ping",
		"data":  param,
	}
	data, _ := json.Marshal(send)
	_, err = conn.Write([]byte(data))
	if err != nil {
		o.ReturnError(-1, err.Error())
		return
	}
	buf := make([]byte, 300000)
	length, err := conn.Read(buf)
	if err != nil {
		o.ReturnError(-1, err.Error())
		return
	}
	o.ReturnResult(string(buf[0:length]))
}

func (o *StatisticController) ProjectChart() {
	var (
		response    map[string]interface{} = make(map[string]interface{})
		re          []keepChart
		statusChart []statusChart
		param       chartParam
		where       string
	)
	if err := o.ParseFormValid(&param); err != nil {
		return
	}

	influxInstance := services.GetInfluxdb()
	db := influxInstance.Table("request_log_" + param.ProjectId).Field("count(response_time),min(response_time),max(response_time),floor(mean(response_time)) AS mean,median(response_time) AS median,sum(status) AS error_num")
	if param.TimeGroup == "" {
		param.TimeGroup = "1"
	}
	if where == "" && param.StartTime == "" {
		where = fmt.Sprintf("time > now() - %dm", services.InterfaceToInt(param.TimeGroup)*60)
	}
	paserWhere(param, &where)
	db.Where(where).GroupBy(fmt.Sprintf("time(%sm)", param.TimeGroup)).Find(&re)
	db.Field("count(type)").Where(where).GroupBy("response_status").Find(&statusChart)
	response["data"] = re
	response["status_chart"] = statusChart
	o.ReturnResult(response)
}

func paserWhere(param chartParam, where *string) {
	if param.StartTime != "" {
		*where = fmt.Sprintf("time > '%s' AND time < '%s'", param.StartTime, param.OverTime)
	}
	if param.RequestNumber != "" {
		*where += fmt.Sprintf(" AND request_number =  '%s'", strings.Trim(param.RequestNumber, " "))
	}
	if param.RequestHost != "" {
		*where += fmt.Sprintf(" AND request_host =  '%s'", param.RequestHost)
	}
	if param.RequestPath != "" {
		*where += fmt.Sprintf(" AND request_path =  '%s'", param.RequestPath)
	}
	if param.KeepServerIp != "" {
		*where += fmt.Sprintf(" AND keep_server_ip =  '%s'", param.KeepServerIp)
	}
	if param.ErrorStatus != "" {
		*where += fmt.Sprintf(" AND error_status='%s'", param.ErrorStatus)
	}
	*where = strings.TrimLeft(*where, " AND")
}

type requestLog struct {
	Time           string      `json:"time"`
	Type           string      `json:"type"`
	RequestHost    string      `json:"request_host"`
	RequestPath    string      `json:"request_path"`
	KeepServerIp   string      `json:"keep_server_ip"`
	ErrorStatus    string      `json:"error_status"`
	ResponseStatus string      `json:"response_status"`
	RequestNumber  string      `json:"request_number"`
	ResponseTime   json.Number `json:"response_time"`
	ResponseData   string      `json:"response_data"`
	ResponseHeader string      `json:"response_header"`
	NetworkCode    string      `json:"network_code"`
	ErrorMsg       string      `json:"error_msg"`
}

func (o *StatisticController) ErrorLogs() {
	var (
		response map[string]interface{} = make(map[string]interface{})
		re       []requestLog
		param    chartParam
		where    string
		count    int64
	)
	if err := o.ParseFormValid(&param); err != nil {
		return
	}
	if param.Limit == 0 {
		param.Limit = 20
	}
	if param.Page == 0 {
		param.Page = 1
	}

	influxInstance := services.GetInfluxdb()
	paserWhere(param, &where)
	fmt.Println("where:", where)
	influxInstance.Table("request_log_"+param.ProjectId).Where(where).Count("type", &count).Limit(param.Limit).Page(param.Page).OrderBy("time DESC").Find(&re)

	response["data"] = re
	response["total"] = count
	response["limit"] = param.Limit
	response["page"] = param.Page
	o.ReturnResult(response)
}

func (o *StatisticController) ProjectDetail() {
	id, _ := o.GetInt("project_id")
	re, err := models.GetProjectDetail(id)
	if err != nil {
		o.ReturnError(-1, err.Error())
		return
	}
	o.ReturnResult(re)
}
