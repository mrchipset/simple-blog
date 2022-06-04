package main

import (
	"blog/blog"
)

func main() {
	db := blog.GetDBConnection()
	blog.MigrateDB(db)
	r := blog.CreateServer()
	r.Run(":8080")
}
