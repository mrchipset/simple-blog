package blog

import (
	"html/template"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/thinkerou/favicon"
)

func CreateServer() *gin.Engine {
	r := gin.Default()
	r.Static("/css", "./css")
	r.Static("/js", "./js")
	r.Static("/resources", "./resources")
	r.Use(favicon.New("./resources/favicon.ico"))
	r.SetFuncMap((template.FuncMap{
		"PrintContent":   PrintContent,
		"Inc":            Inc,
		"PrepareContent": PrepareContent,
	}))

	trustedProxies := os.Getenv("TRUSTED_PROXIES")
	proxies := strings.Split(trustedProxies, ";")
	r.SetTrustedProxies(proxies)
	r.LoadHTMLGlob("templates/**/*")
	view := r.Group("/")
	{
		view.GET("/", MainViewHandle)
		view.GET("/post/:uuid", PostViewHandle)
		view.GET("/page/:uuid", PageViewHandle)
	}

	editor := r.Group("/editor")
	{
		editor.GET("/", AuthMiddware(), PostListViewHandle)
		editor.GET("/login", LoginViewHandle)
		editor.POST("/login", LoginViewAPIHandle)
		editor.GET("/compose/:uuid", AuthMiddware(), EditorViewHandle)
		editor.POST("/preview", AuthMiddware(), PreviewAPIHandle)
		editor.POST("/save/:uuid", AuthMiddware(), SavePostAPIHandle)
		editor.POST("/publish/:uuid", AuthMiddware(), PublishPostAPIHandle)
		editor.POST("/create", AuthMiddware(), CreateNewPostAPIHandle)
		editor.POST("/delete/:uuid", AuthMiddware(), DeletePostAPIHandle)
	}

	return r
}
