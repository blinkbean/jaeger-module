package jaegergormv2

import (
	jaegerModule "github.com/blinkbean/jaeger-module"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
)

var dataSource = jaegerModule.DATASOURCE
var serviceName = "jaeger_gorm_v2"

func TestGormWithoutOpenTracing(t *testing.T) {
	closer := jaegerModule.InitJaeger(serviceName)
	defer closer.Close()
	assert := assert.New(t)
	db, err := gorm.Open(mysql.Open(dataSource))
	assert.Equal(err, nil, "open err")
	create(db, assert)
	update(db, assert)
	query(db, assert)
	row(db, assert)
	raw(db, assert)
	delete(db, assert)
}

func TestGormOpenTracing(t *testing.T) {
	closer := jaegerModule.InitJaeger(serviceName)
	defer closer.Close()
	assert := assert.New(t)
	db, err := gorm.Open(mysql.Open(dataSource))
	db.Use(opentracingPlugin{})
	assert.Equal(err, nil, "open err")
	create(db, assert)
	update(db, assert)
	query(db, assert)
	row(db, assert)
	raw(db, assert)
	delete(db, assert)
}

type Jaeger struct {
	Id       int64  `json:"id"`
	JaegerId int64  `json:"jaeger_id"`
	Text     string `json:"text"`
}

func (j Jaeger) TableName() string {
	return "jaeger"
}

var data = []Jaeger{{1, 1, "1"}, {2, 2, "2"}, {3, 3, "3"}}

func create(db *gorm.DB, assert *assert.Assertions) {
	for _, v := range data {
		err := db.Model(&Jaeger{}).Create(&v).Error
		assert.Equal(err, nil, "create err")
	}
}

func update(db *gorm.DB, assert *assert.Assertions) {
	for _, v := range data {
		m := map[string]interface{}{"text": v.Text + v.Text}
		err := db.Model(&Jaeger{}).Where("id = ?", v.Id).UpdateColumns(m).Error
		assert.Equal(err, nil, "update err")
	}
}

func query(db *gorm.DB, assert *assert.Assertions) {
	for _, v := range data {
		var j Jaeger
		err := db.Model(&Jaeger{}).Where("id = ?", v.Id).First(&j).Error
		assert.Equal(err, nil, "get err")
		assert.Equal(v.JaegerId, j.JaegerId, "get data err")
	}
}

func row(db *gorm.DB, assert *assert.Assertions) {
	for _, v := range data {
		row := db.Model(&Jaeger{}).Select("jaeger_id").Where("id = ?", v.Id).Row()
		var coterieId int64
		row.Scan(&coterieId)
		assert.Equal(coterieId, v.Id, "row err")
	}

	rows, err := db.Model(&Jaeger{}).Order("id").Rows()
	defer rows.Close()
	assert.Equal(err, nil, "rows err")
	var j Jaeger
	var i int
	for rows.Next() {
		db.ScanRows(rows, &j)
		assert.Equal(j.Id, data[i].Id, "get rows err")
		i++
	}
}

func raw(db *gorm.DB, assert *assert.Assertions) {
	var j Jaeger
	err := db.Raw("select * from jaeger where id = 1").Scan(&j).Error
	assert.Equal(err, nil, "raw get err")
	assert.Equal(j.JaegerId, int64(1), "get raw data err")
}

func delete(db *gorm.DB, assert *assert.Assertions) {
	for _, v := range data {
		err := db.Delete(v).Error
		assert.Equal(err, nil, "delete err")
	}
}
