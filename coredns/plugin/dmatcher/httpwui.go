package dmatcher

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type WUI struct {
	Conf *DConf
	//srv    *http.Server
	router *gin.Engine
}

func (wui *WUI) startWUI() {
	gin.SetMode(gin.ReleaseMode)
	wui.router = gin.Default()

	wui.router.Static("/static", "./www_static")
	wui.router.LoadHTMLGlob("templates/*")
	//	wui.router.GET("/", wui.wuiGinRootPage)
	wui.router.StaticFile("/", "./www_static/index.html")
	wui.router.GET("/query/:fqdn", wui.wuiGinWWWQuery)
	wui.router.GET("/query/", wui.wuiGinWWWQueryRoot)
	wui.router.GET("/api/query/:fqdn", wui.wuiGinQueryDomains)
	wui.router.GET("/api/query/", wui.wuiGinQueryDomainsRoot)
	wui.router.GET("/api/add/:fqdn", wui.wuiGinAddDomain)
	wui.router.GET("/api/del/:fqdn", wui.wuiGinDelDomain)
	wui.router.DELETE("/api/del/:fqdn", wui.wuiGinDelDomain)
	wui.router.GET("/api/load", func(ctx *gin.Context) {
		wui.Conf.storageInstance.Load()
	})
	wui.router.GET("/api/save", func(ctx *gin.Context) {
		wui.Conf.storageInstance.Save()
	})
	go wui.router.Run(":" + wui.Conf.Port)

}

func (wui *WUI) wuiGinWWWQueryRoot(c *gin.Context) {
	list, _ := wui.Conf.storageInstance.GetDomainList(".")
	if len(list) > 100 {
		list = list[:100]
	}
	//fmt.Println(list)
	c.HTML(http.StatusOK, "table.tmpl", gin.H{
		"Domains": list,
	})
}

func (wui *WUI) wuiGinWWWQuery(c *gin.Context) {
	list, _ := wui.Conf.storageInstance.GetDomainList(c.Param("fqdn"))
	if len(list) > 100 {
		list = list[:100]
	}
	//fmt.Println(list)
	c.HTML(http.StatusOK, "table.tmpl", gin.H{
		"Domains": list,
	})
}

func (wui *WUI) wuiGinQueryDomainsRoot(c *gin.Context) {

	list, _ := wui.Conf.storageInstance.GetDomainList(".")
	var str string
	for _, s := range list {
		str = str + "\n" + s
	}

	c.String(http.StatusOK, "%s", str)
}

func (wui *WUI) wuiGinQueryDomains(c *gin.Context) {

	list, _ := wui.Conf.storageInstance.GetDomainList(c.PostForm("suffix"))
	var str string
	for _, s := range list {
		str = str + "\n" + s
	}

	c.String(http.StatusOK, "%s", str)
}
func (wui *WUI) wuiGinAddDomain(c *gin.Context) {
	wui.Conf.storageInstance.AddDomain(c.Param("fqdn"))
	wui.Conf.storageInstance.Save()
	go wui.notifyRemoteInstance("/api/add/", c.Param("fqdn"))
	/*list, _ := wui.Conf.storageInstance.GetDomainList(c.Param("fqdn"))
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"Domains": list,
	})*/
}
func (wui *WUI) wuiGinDelDomain(c *gin.Context) {
	wui.Conf.storageInstance.DelDomain(c.Param("fqdn"))
	wui.Conf.storageInstance.Save()
	go wui.notifyRemoteInstance("/api/del/", c.Param("fqdn"))
	c.HTML(http.StatusOK, "", "")
	/*list, _ := wui.Conf.storageInstance.GetDomainList(c.Param("fqdn"))
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"Domains": list,
	})*/
}

func (wui *WUI) notifyRemoteInstance(endpoint string, data string) {
	for _, addr := range wui.Conf.notifyOther {
		requestURL := fmt.Sprintf("http://%s%s%s", addr, endpoint, data)
		_, err := http.Get(requestURL)
		if err != nil {
			fmt.Printf("error making http request: %s\n", err)
		}
	}
}
