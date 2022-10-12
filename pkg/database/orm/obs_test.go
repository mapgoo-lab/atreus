package orm

import (
	"context"
	"github.com/mapgoo-lab/atreus/pkg/conf/env"
	"github.com/mapgoo-lab/atreus/pkg/net/trace"
	"github.com/mapgoo-lab/atreus/pkg/net/trace/mocktrace"
	"github.com/mapgoo-lab/atreus/pkg/net/trace/zipkin"
	"gorm.io/gorm"
	"testing"
	"gorm.io/driver/sqlite"
	"time"
)

type Product struct {
	gorm.Model
	Code  string
	Price uint
}


func TestOrmObs(t *testing.T)  {

	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db.Use(&ObsPlugin{})

	db.AutoMigrate(&Product{})


	tr := &mocktrace.MockTrace{}
	tr.Inject(nil, nil, nil)
	tr.Extract(nil, nil)

	root := tr.New("test")

	session := db.WithContext(trace.NewContext(context.Background(), root))

	session.Create(&Product{Code: "D42", Price: 100})

	//Read
	var product Product
	session.First(&product, 1) // find product with integer primary key
	session.First(&product, "code = ?", "D42") // find product with code D42

	// Update - update product's price to 200
	session.Model(&product).Update("Price", 200)
	// Update - update multiple fields
	session.Model(&product).Updates(Product{Price: 200, Code: "F42"}) // non-zero fields
	session.Model(&product).Updates(map[string]interface{}{"Price": 200, "Code": "F42"})

	// Delete - delete product
	session.Delete(&product, 1)
}

func TestOrmObsZipkin(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db.Use(&ObsPlugin{})

	db.AutoMigrate(&Product{})

	config := &zipkin.Config{
		Endpoint: "http://192.168.100.16:9411/api/v2/spans",
		BatchSize: 3,
		DisableSample: false,
	}

	env.AppID = "TestOrmObsZipkin"

	zipkin.Init(config)

	session := db.WithContext(trace.NewContext(context.Background(), trace.New("TestOrmObsZipkin")))

	session.Create(&Product{Code: "D42", Price: 100})

	//Read
	var product Product
	session.First(&product, 1) // find product with integer primary key
	session.First(&product, "code = ?", "D42") // find product with code D42

	// Update - update product's price to 200
	session.Model(&product).Update("Price", 200)
	// Update - update multiple fields
	session.Model(&product).Updates(Product{Price: 200, Code: "F42"}) // non-zero fields
	session.Model(&product).Updates(map[string]interface{}{"Price": 200, "Code": "F42"})

	// Delete - delete product
	session.Delete(&product, 1)

	time.Sleep(30 * time.Second)
}