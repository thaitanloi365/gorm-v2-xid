package main

import (
	"fmt"
	"reflect"

	"github.com/rs/xid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type Model struct {
	ID        string `gorm:"primary_key" json:"id"`
	CreatedAt int64  `gorm:"autoCreateTime" json:"-"`
	UpdatedAt int64  `gorm:"autoUpdateTime" json:"-"`
}
type Lot struct {
	Model
	Name  string  `json:"name"`
	Spots []*Spot `gorm:"foreignKey:LotID;references:ID" json:"spots"`
}

type Spot struct {
	Model
	LotID string `json:"-"`

	Name string `json:"name"`
}

func setField(field *schema.Field, rv reflect.Value) {
	if _, isZero := field.ValueOf(rv); isZero {
		var xid = xid.New().String()
		field.Set(rv, xid)
	}
}

func main() {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db = db.Debug()

	db.AutoMigrate(&Lot{}, &Spot{})

	db.Callback().Create().Before("gorm:save_before_associations").Register("app:update_xid_when_create", func(db *gorm.DB) {

		fmt.Println(db.Statement.ReflectValue.Kind())
		switch db.Statement.ReflectValue.Kind() {
		case reflect.Slice, reflect.Array:
			for i := 0; i < db.Statement.ReflectValue.Len(); i++ {
				var rv = db.Statement.ReflectValue.Index(i)
				var field = db.Statement.Schema.LookUpField("ID")
				setField(field, rv)
			}
		case reflect.Struct:
			var field = db.Statement.Schema.LookUpField("ID")
			setField(field, db.Statement.ReflectValue)
		}

	})

	var id1 = xid.New().String()
	var id2 = xid.New().String()
	fmt.Println(id1, id2)
	var lots = []*Lot{
		{
			Model: Model{
				ID: id1,
			},
			Name: "Lot 1",
			Spots: []*Spot{
				{
					Model: Model{
						ID: id2,
					},
					Name: "Spot 11",
				},
				{
					Name: "Spot 12",
				},
			},
		},
		{
			Name: "Lot 2",
			Spots: []*Spot{
				{
					Name: "Spot 21",
				},
				{
					Name: "Spot 22",
				},
			},
		},
	}

	db.Save(&lots)
}
