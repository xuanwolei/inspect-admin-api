/*
 * @Author: your name
 * @Date: 2020-04-01 17:40:04
 * @LastEditTime: 2020-09-17 20:49:45
 * @LastEditors: ybc
 * @Description: In User Settings Edit
 * @FilePath: \storyMarketingd:\go1.13.5\path\src\inspect\services\influxdb.go
 */
package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod
	client "github.com/influxdata/influxdb1-client/v2"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	InfluxDbInstance *InfluxInstance
	lock             sync.Mutex
)

type InfluxInstance struct {
	Influx         client.Client
	InfluxResponse *client.Response
	QueryMap       *InfluxQuery
	Scope          *Scope
	TagName        string
}
type InfluxQuery struct {
	DbName    string
	TableName string
	Field     string
	Where     string
	Limit     int
	Offset    int
	Page      int
	Group     string
	Order     string
	Sql       string
}

type QueryResult []map[string]interface{}

type Scope struct {
	TopValue reflect.Value
	Value    reflect.Value
	Type     reflect.Type
}

type InfluxConfig struct {
	HttpConfig client.HTTPConfig
	DbName     string
	TagName    string
}

func GetInfluxdb() *InfluxInstance {
	config := &InfluxConfig{
		HttpConfig: client.HTTPConfig{
			Addr:    beego.AppConfig.String("influx.addr"),
			Timeout: time.Minute * 10,
		},
		DbName: beego.AppConfig.String("influx.database"),
	}
	db, err := GetInfluxdbInstance(config)
	if err != nil {
		panic(err.Error())
	}
	return db
}

func GetInfluxdbInstance(config *InfluxConfig) (*InfluxInstance, error) {
	// lock.Lock()
	// defer lock.Unlock()
	// if InfluxDbInstance != nil {
	// 	//ping通无需重新加载
	// 	if _, _, err := InfluxDbInstance.Influx.Ping(time.Second * 5); err == nil {
	// 		return InfluxDbInstance, nil
	// 	}
	// }
	cli, err := NewInfluxdbClient(config.HttpConfig)
	if err != nil {
		return nil, err
	}
	InfluxDbInstance := &InfluxInstance{
		Influx: cli,
		QueryMap: &InfluxQuery{
			DbName: config.DbName,
		},
		TagName: "xw",
	}

	return InfluxDbInstance, nil
}

func NewInfluxdbClient(config client.HTTPConfig) (client.Client, error) {

	influxdbClient, err := client.NewHTTPClient(config)
	if err != nil {
		return nil, errors.New("Error creating InfluxDB Client: " + err.Error())
	}

	return influxdbClient, nil
}

func (this *InfluxInstance) NewScope(value interface{}) *InfluxInstance {
	getType := reflect.TypeOf(value).Elem().Elem()
	getValue := reflect.New(getType)
	if getValue.Kind() == reflect.Ptr {
		getValue = getValue.Elem()
	}
	this.Scope = &Scope{
		TopValue: reflect.ValueOf(value),
		Value:    getValue,
		Type:     getType,
	}
	return this
}

func (this *InfluxInstance) SelectDb(dbName string) *InfluxInstance {
	this.QueryMap.DbName = dbName
	return this
}

func (this *InfluxInstance) Table(table string) *InfluxInstance {
	this.QueryMap.TableName = table
	return this
}

func (this *InfluxInstance) Field(value string) *InfluxInstance {
	this.QueryMap.Field = value
	return this
}

func (this *InfluxInstance) Where(where string) *InfluxInstance {
	this.QueryMap.Where = where
	return this
}

func (this *InfluxInstance) OrderBy(order string) *InfluxInstance {
	this.QueryMap.Order = order
	return this
}

func (this *InfluxInstance) GroupBy(group string) *InfluxInstance {
	this.QueryMap.Group = group
	return this
}

func (this *InfluxInstance) Limit(limit int) *InfluxInstance {
	this.QueryMap.Limit = limit
	return this
}

func (this *InfluxInstance) Offset(offset int) *InfluxInstance {
	this.QueryMap.Offset = offset
	return this
}

//分页、需要先设置limit
func (this *InfluxInstance) Page(page int) *InfluxInstance {
	this.QueryMap.Page = page
	this.Offset((page - 1) * this.QueryMap.Limit)
	return this
}

func (this *InfluxInstance) Count(field string, count *int64) *InfluxInstance {
	this.Field(fmt.Sprintf("count(%s)", field))
	res, err := this.QueryWithResult()
	if err != nil {
		return this
	}
	if len(res) == 0 {
		*count = 0
		return this
	}

	*count, _ = res[0]["count"].(json.Number).Int64()
	this.Field("")
	return this
}

func (this *InfluxInstance) Find(value interface{}) error {
	res, err := this.NewScope(value).QueryWithResult()
	if err != nil {
		return err
	}
	fmt.Println("sql:", this.GetLastSql())
	this.Mapping(res).resetParam()
	return nil
}

func (this *InfluxInstance) resetParam() {
	this.QueryMap = &InfluxQuery{
		DbName:    this.QueryMap.DbName,
		TableName: this.QueryMap.TableName,
	}
}

func (this *InfluxInstance) Mapping(res QueryResult) *InfluxInstance {
	var (
		reValue  reflect.Value
		allValue reflect.Value = this.Scope.TopValue.Elem()
	)
	for _, value := range res {
		reValue = this.Scope.Value
		for key, data := range value {
			key = StrUnderlineToUpper(key)
			valid := reValue.FieldByName(key).IsValid()
			if data != nil && valid {
				reValue.FieldByName(key).Set(reflect.ValueOf(data))
			}
		}
		allValue = reflect.Append(allValue, reValue)
	}
	this.Scope.TopValue.Elem().Set(allValue)
	return this
}

func (this *InfluxInstance) QueryWithResult() (QueryResult, error) {
	this.parseQuerySql()
	re, err := this.Query(client.NewQuery(this.QueryMap.Sql, this.QueryMap.DbName, ""))

	if err != nil {
		return nil, err
	}

	return re, nil
}

func (this *InfluxInstance) Query(q client.Query) (QueryResult, error) {
	var dataMap QueryResult
	response, err := this.Influx.Query(q)
	if err != nil && response.Error() != nil {
		return nil, err
	}
	this.InfluxResponse = response
	for _, result := range response.Results {

		for _, row := range result.Series {
			for _, value := range row.Values {
				mp := make(map[string]interface{}, len(row.Columns))
				for i := 0; i < len(row.Columns); i++ {
					mp[row.Columns[i]] = value[i]
				}
				for k, v := range row.Tags {
					mp[k] = v
				}
				dataMap = append(dataMap, mp)
			}
		}
	}

	return dataMap, nil
}

func (this *InfluxInstance) parseQuerySql() *InfluxInstance {
	var (
		field  string
		limit  string
		where  string
		group  string
		offset string
	)
	if this.QueryMap.Field == "" {
		for i := 0; i < this.Scope.Value.NumField(); i++ {
			fieldType := this.Scope.Value.Type().Field(i)
			if fieldType.Tag.Get(this.TagName) != "" {
				field += fieldType.Tag.Get(this.TagName) + ","
			} else {
				field += StrToUnderlineWithLower(fieldType.Name) + ","
			}
		}
		this.QueryMap.Field = strings.TrimRight(field, ",")
	}
	if this.QueryMap.TableName == "" {
		this.QueryMap.TableName = this.Scope.Type.Name()
	}

	if this.QueryMap.Limit != 0 {
		limit = fmt.Sprintf("LIMIT %d", this.QueryMap.Limit)
	}
	if this.QueryMap.Offset > 0 {
		offset = fmt.Sprintf("OFFSET %d", this.QueryMap.Offset)
	}
	if this.QueryMap.Where != "" {
		where = fmt.Sprintf("WHERE %s", this.QueryMap.Where)
	}

	if this.QueryMap.Group != "" {
		group = fmt.Sprintf("GROUP BY %s", this.QueryMap.Group)
	} else if this.QueryMap.Order != "" {
		group = fmt.Sprintf("ORDER BY %s", this.QueryMap.Order)
	}
	this.QueryMap.Sql = fmt.Sprintf("SELECT %s FROM %s %s %s %s %s %s", this.QueryMap.Field, this.QueryMap.TableName, where, group, limit, offset, "tz('Asia/Shanghai')")

	return this
}

func (this *InfluxInstance) GetLastSql() string {
	return this.QueryMap.Sql
}

func formatAtom(v reflect.Value) string {
	fmt.Println(v.String())
	switch v.Kind() {
	case reflect.Invalid:
		return "invalid"
	case reflect.Int, reflect.Int8, reflect.Int16,
		reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(v.Uint(), 10)
	// ...floating-point and complex cases omitted for brevity...
	case reflect.Bool:
		return strconv.FormatBool(v.Bool())
	case reflect.String:
		return strconv.Quote(v.String())
	case reflect.Chan, reflect.Func, reflect.Ptr, reflect.Slice, reflect.Map:
		return v.Type().String() + " 0x" +
			strconv.FormatUint(uint64(v.Pointer()), 16)
	default: // reflect.Array, reflect.Struct, reflect.Interface
		return v.Type().String() + " value"
	}
}

func StrFirstToUpper(str string) string {
	if len(str) < 1 {
		return ""
	}
	strArry := []rune(str)
	if strArry[0] >= 97 && strArry[0] <= 122 {
		strArry[0] -= 32
	}
	return string(strArry)
}

//
func StrUnderlineToUpper(str string) string {
	if len(str) < 1 {
		return ""
	}
	strArry := strings.Split(str, "_")
	var newStr string
	for _, v := range strArry {
		newStr += StrFirstToUpper(v)
	}

	return newStr
}

func StrToUnderlineWithLower(str string) string {
	if len(str) < 1 {
		return ""
	}
	var newStr []rune
	strArry := []rune(str)
	for _, v := range strArry {
		if v < 91 {
			newStr = append(newStr, 95)
			v += 32
		}
		newStr = append(newStr, v)
	}

	return strings.TrimLeft(string(newStr), "_")
}

// q := client.NewQuery("SELECT * FROM keep WHERE time > 1585216763773907098 limit 5", "ybc", "")
// 	if response, err := influxInstance.Influx.Query(q); err == nil && response.Error() == nil {
// 		for _, result := range response.Results {
// 			for _, row := range result.Series {
// 				// fmt.Println("name:", row.Name)
// 				// fmt.Println("tags:", row.Tags)
// 				fmt.Println("Columns:", row.Columns)
// 				// fmt.Println("Values:", row.Values)

// 				for _, value := range row.Values {
// 					fmt.Println("value:", value[1])
// 				}
// 			}
// 		}
// 	}
