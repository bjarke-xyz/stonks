package web

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/bjarke-xyz/stonks/internal/core"
	"github.com/bjarke-xyz/stonks/internal/web/views"
	"github.com/gin-gonic/gin"
)

//go:embed static/*
var static embed.FS

type web struct {
	appContext *core.AppContext
}

func NewWeb(appContext *core.AppContext) *web {
	return &web{appContext: appContext}
}

func (w *web) Route(r *gin.Engine) {
	staticFiles(r, static)
	r.HEAD("/", w.HandleGetIndex)
	r.GET("/", w.HandleGetIndex)
	r.GET("/quote/:symbol", w.HandleGetQuote)
}

func (w *web) getBaseModel(c *gin.Context, title string) views.BaseViewModel {
	var unixBuildTime int64 = 0
	if w.appContext.Config.BuildTime != nil {
		unixBuildTime = w.appContext.Config.BuildTime.Unix()
	} else {
		unixBuildTime = time.Now().Unix()
	}
	// hxRequest := c.Request.Header.Get("HX-Request")
	// includeLayout := hxRequest == "" || hxRequest == "false"
	// log.Println("hxRequest", hxRequest, "includeLayout", includeLayout)
	model := views.BaseViewModel{
		Path:          c.Request.URL.Path,
		UnixBuildTime: unixBuildTime,
		Title:         title + " | stonks",
		FlashInfo:     GetFlashes(c, core.FlashTypeInfo),
		FlashWarn:     GetFlashes(c, core.FlashTypeWarn),
		FlashError:    GetFlashes(c, core.FlashTypeError),
	}
	return model
}

func (w *web) handleError(c *gin.Context, err error) {
	log.Printf("error: %v", err)
	errViewModel := views.ErrViewModel{
		Base:  w.getBaseModel(c, "error"),
		Error: err,
	}
	c.HTML(http.StatusInternalServerError, "", views.Error(errViewModel))
}

func staticFiles(r *gin.Engine, staticFs fs.FS) {
	staticWeb, err := fs.Sub(staticFs, "static")
	if err != nil {
		log.Printf("failed to get fs sub for static: %v", err)
	}
	httpFsStaticWeb := http.FS(staticWeb)
	r.Use(staticCacheMiddleware())
	r.StaticFS("/static", httpFsStaticWeb)
	r.StaticFileFS("/favicon.ico", "./favicon.ico", httpFsStaticWeb)
	r.StaticFileFS("/favicon-16x16.png", "./favicon-16x16.png", httpFsStaticWeb)
	r.StaticFileFS("/favicon-32x32.png", "./favicon-32x32.png", httpFsStaticWeb)
	r.StaticFileFS("/apple-touch-icon.png", "./apple-touch-icon.png", httpFsStaticWeb)
	r.StaticFileFS("/site.webmanifest", "./site.webmanifest", httpFsStaticWeb)

}

func staticCacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/static/js") || strings.HasPrefix(path, "/static/css") {
			c.Header("Cache-Control", "public, max-age=31536000, immutable")
		}
		c.Next()
	}
}
