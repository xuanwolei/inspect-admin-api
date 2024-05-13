/*
 * @Author: yuanbaocheng
 * @Date: 2020-04-28 21:46:53
 * @LastEditors: ybc
 * @LastEditTime: 2020-09-18 14:36:30
 * @Description: file content
 */
package models

import (
	"encoding/json"
	"errors"
	"inspect/services"
)

type Config struct {
	Id       uint   `gorm:"primary_key" json:"id"`
	Name     string `json:"name"`
	HttpHost string `json:"http_host"`
	Paths    string `json:"paths"`
}

type Configs struct {
	Id       uint   `gorm:"primary_key" json:"id"`
	Hosts    string `json:"hosts"`
	Type     string `json:"type"`
	Level    uint   `json:"level"`
	Timeout  uint   `json:"timeout"`
	State    uint   `json:"state"`
	Desc     string `json:"desc"`
	Name     string `json:"name"`
	HttpHost string `json:"http_host"`
	Paths    string `json:"paths"`
	Phone    string `json:"phone"`
}

type Projects struct {
	Data  []Configs `json:"data"`
	Total int       `json:"total"`
	Limit int       `json:"limit"`
	Page  int       `json:"page"`
}

type PageData struct {
	Total int `json:"total"`
	Limit int `json:"limit"`
	Page  int `json:"page"`
}

const (
	PAGE_LITMIT int = 10
)

func GetProjects(where interface{}, page int) (*Projects, error) {

	var re []Configs
	var count int
	res := services.MysqlInstance.Table("config").Where("state = 1").Where(where).Count(&count).Offset((page - 1) * PAGE_LITMIT).Order("id DESC").Limit(PAGE_LITMIT).Find(&re)
	data := &Projects{
		Data:  re,
		Total: count,
		Limit: PAGE_LITMIT,
		Page:  page,
	}
	if res.RecordNotFound() {
		return data, errors.New("没有找到结果")
	}
	return data, nil
}

func GetProjectDetail(id int) (Configs, error) {
	var re Configs
	where := make(map[string]interface{})
	where["id"] = id
	res := services.MysqlInstance.Table("config").Where(where).First(&re)

	if res.RecordNotFound() {
		return re, errors.New("没有找到结果")
	}
	return re, nil
}

func GetProjectDetailFormat(id int) (map[string]interface{}, error) {
	var (
		uid      uint
		name     string
		httpHost string
		paths    string
	)
	where := make(map[string]interface{})
	result := make(map[string]interface{})
	where["id"] = id
	row := services.MysqlInstance.Table("config").Select("id,name,http_host,paths").Where(where).Row()
	// row.Scan(&uid, &name, &httpHost, &paths)
	row.Scan(&uid, &name, &httpHost, &paths)
	// if res.RecordNotFound() {
	// 	return result, errors.New("没有找到结果")
	// }
	var it interface{}
	json.Unmarshal([]byte(paths), &it)
	result["id"] = uid
	result["name"] = name
	result["http_host"] = httpHost
	result["paths"] = it

	return result, nil
}
