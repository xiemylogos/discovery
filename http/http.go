package http

import (
	"net/http"

	"github.com/bilibili/discovery/conf"
	"github.com/bilibili/discovery/discovery"
	"github.com/bilibili/kratos/pkg/log"
	bm "github.com/bilibili/kratos/pkg/net/http/blademaster"
	"github.com/pkg/errors"
)

var (
	dis          *discovery.Discovery
	protected    = true
	errProtected = errors.New("discovery in protect mode and only support register")
)

// Init init http
func Init(c *conf.Config, s *discovery.Discovery, isTls bool, certFile, keyFile string) {
	dis = s
	engineInner := bm.DefaultServer(c.HTTPServer)
	innerRouter(engineInner)
	if !isTls {
		if err := engineInner.Start(); err != nil {
			log.Error("bm.DefaultServer error(%v)", err)
			panic(err)
		}
	} else {
		go func() {
			if err := engineInner.RunTLS(c.HTTPServer.Addr, "", ""); err != nil {
				if errors.Cause(err) == http.ErrServerClosed {
					log.Info("RunTls: server closed")
					return
				}
				panic(errors.Wrapf(err, "RunTLS: engine.ListenServer %v)", c.HTTPServer.Addr))
			}
		}()
	}

}

// innerRouter init local router api path.
func innerRouter(e *bm.Engine) {
	group := e.Group("/discovery")
	{
		group.POST("/register", register)
		group.POST("/renew", renew)
		group.POST("/cancel", cancel)
		group.GET("/fetch/all", initProtect, fetchAll)
		group.GET("/fetchapp", initProtect, fetchApp)
		group.GET("/fetchapps", initProtect, fetchApps)
		group.GET("/fetch", initProtect, fetch)
		group.GET("/fetchs", initProtect, fetchs)
		group.GET("/poll", initProtect, poll)
		group.GET("/polls", initProtect, polls)
		//manager
		group.POST("/set", set)
		group.GET("/nodes", initProtect, nodes)
	}
}

func initProtect(ctx *bm.Context) {
	if dis.Protected() {
		ctx.JSON(nil, errProtected)
		ctx.AbortWithStatus(503)
	}
}
