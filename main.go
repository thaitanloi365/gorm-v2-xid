package main

import (
	"fmt"
	"reflect"

	"github.com/rs/xid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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

func main() {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db = db.Debug()

	db.AutoMigrate(&Lot{}, &Spot{})

	db.Callback().Create().Before("gorm:save_before_associations").Register("app:update_xid_when_create", func(db *gorm.DB) {
		var field = db.Statement.Schema.LookUpField("ID")
		if field != nil {
			if v, isZero := field.ValueOf(db.Statement.ReflectValue); isZero {
				if _, ok := v.(string); ok {
					fmt.Println("****** kind", db.Statement.ReflectValue.Kind())
					switch db.Statement.ReflectValue.Kind() {
					case reflect.Slice, reflect.Array:
						for i := 0; i < db.Statement.ReflectValue.Len(); i++ {
							var xid = xid.New().String()
							field.Set(db.Statement.ReflectValue.Index(i), xid)
						}
					case reflect.Struct:
						var xid = xid.New().String()
						field.Set(db.Statement.ReflectValue, xid)
					}
				}
			}
		}

	})

	var lots = []*Lot{
		{
			Name: "Lot 1",
			Spots: []*Spot{
				{
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

	for _, lot := range lots {
		db.Create(&lot)
	}
}
