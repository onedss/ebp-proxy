package routers

import (
	"fmt"
	"github.com/gin-contrib/pprof"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/onedss/ebp-proxy/cors"
	"github.com/onedss/ebp-proxy/db"
	"github.com/onedss/ebp-proxy/mylog"
	"github.com/onedss/ebp-proxy/sessions"
	validator "gopkg.in/go-playground/validator.v8"
	"log"
	"mime"
	"net/http"
	"path/filepath"
)

/**
 * @apiDefine sys API接口
 */
var Router *gin.Engine

func init() {
	mime.AddExtensionType(".svg", "image/svg+xml")
	mime.AddExtensionType(".m3u8", "application/vnd.apple.mpegurl")
	// mime.AddExtensionType(".m3u8", "application/x-mpegurl")
	mime.AddExtensionType(".ts", "video/mp2t")
	// prevent on Windows with Dreamware installed, modified registry .css -> application/x-css
	// see https://stackoverflow.com/questions/22839278/python-built-in-server-not-loading-css
	mime.AddExtensionType(".css", "text/css; charset=utf-8")

	gin.DisableConsoleColor()
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = mylog.GetLogWriter()
}

type APIHandler struct {
	RestartChan chan bool
}

var API = &APIHandler{
	RestartChan: make(chan bool),
}

func Errors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		for _, err := range c.Errors {
			switch err.Type {
			case gin.ErrorTypeBind:
				switch err.Err.(type) {
				case validator.ValidationErrors:
					errs := err.Err.(validator.ValidationErrors)
					for _, err := range errs {
						sec := mylog.Conf().Section("localize")
						field := sec.Key(err.Field).MustString(err.Field)
						tag := sec.Key(err.Tag).MustString(err.Tag)
						c.AbortWithStatusJSON(http.StatusBadRequest, fmt.Sprintf("%s %s", field, tag))
						return
					}
				default:
					log.Println(err.Err.Error())
					c.AbortWithStatusJSON(http.StatusBadRequest, "Inner Error")
					return
				}
			}
		}
	}
}

func Init() (err error) {
	Router = gin.New()
	pprof.Register(Router)
	// Router.Use(gin.Logger())
	Router.Use(gin.Recovery())
	Router.Use(Errors())
	Router.Use(cors.Default())

	tokenTimeout := mylog.Conf().Section("http").Key("token_timeout").MustInt(7 * 86400)
	webRoot := mylog.Conf().Section("http").Key("www_root").MustString("www")

	store := sessions.NewGormStoreWithOptions(db.SQLite, sessions.GormStoreOptions{
		TableName: "t_sessions",
	}, []byte("OneDss@2018"))
	store.Options(sessions.Options{HttpOnly: true, MaxAge: tokenTimeout, Path: "/"})
	sessionHandle := sessions.Sessions("token", store)

	{
		wwwDir := filepath.Join(mylog.DataDir(), webRoot)
		log.Println("www root -->", wwwDir)
		Router.Use(static.Serve("/", static.LocalFile(wwwDir, true)))
	}

	{
		api := Router.Group("/api/v1").Use(sessionHandle)
		api.GET("/restart", API.Restart)
	}
	return
}
