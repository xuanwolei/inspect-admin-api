/*
 * @Author: yuanbaocheng
 * @Date: 2020-04-19 21:26:29
 * @LastEditors: ybc
 * @LastEditTime: 2020-09-15 19:34:45
 * @Description: file content
 */
package services

import (
	"fmt"

	"github.com/astaxie/beego"
	"github.com/garyburd/redigo/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/xuanwolei/goutils"
)

var (
	RedisInstance redis.Conn
	MysqlInstance *gorm.DB
	OssInstance   *goutils.AliOss
)

func init() {
	var err error
	MysqlInstance, err = NewMysqlInstance()
	if err != nil {
		panic(err)
	}
}

func NewMysqlInstance() (*gorm.DB, error) {
	db, err := gorm.Open("mysql", fmt.Sprintf("%s:%s@(%s:%s)/%s?charset=%s&parseTime=True&loc=Local", beego.AppConfig.String("mysql.user"), beego.AppConfig.String("mysql.password"), beego.AppConfig.String("mysql.host"), beego.AppConfig.String("mysql.port"), beego.AppConfig.String("mysql.db"), beego.AppConfig.String("mysql.charset")))
	if err != nil {
		return nil, err
	}
	//关闭自动加s
	db.SingularTable(true)
	//设置表前缀
	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return beego.AppConfig.String("mysql.prefix") + defaultTableName
	}
	//设置数据库连接池参数
	db.DB().SetMaxOpenConns(50) //设置数据库连接池最大连接数
	db.DB().SetMaxIdleConns(20) //连接池最大允许的空闲连接数，如果没有sql任务需要执行的连接数大于20，超过的连接会被连接池关闭。
	return db, nil
}
