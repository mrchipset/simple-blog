package blog

import (
	"sync"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB
var gLock sync.Mutex

func CreateDBConnection() *gorm.DB {
	gLock.Lock()
	defer gLock.Unlock()

	var err error
	db, err = gorm.Open(postgres.New(postgres.Config{
		DSN:                  "user=postgres password=postgres dbname=postgres port=5432 host=localhost sslmode=disable TimeZone=Asia/Shanghai",
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), &gorm.Config{})

	if err != nil {
		panic(err)
	}
	return db
}

func GetDBConnection() *gorm.DB {
	if db == nil {
		CreateDBConnection()
	}

	return db
}

func MigrateDB(session *gorm.DB) {
	session.AutoMigrate(&Post{})
	session.AutoMigrate(&PostContent{})
}
