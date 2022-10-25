package base

import (
	"net/http"

	"github.com/allanpk716/ChineseSubFinder/pkg/local_http_proxy_server"
	"github.com/allanpk716/ChineseSubFinder/pkg/settings"

	"github.com/allanpk716/ChineseSubFinder/pkg/tmdb_api"
	"github.com/allanpk716/ChineseSubFinder/pkg/types/backend"
	"github.com/gin-gonic/gin"
)

func (cb *ControllerBase) CheckTmdbApiHandler(c *gin.Context) {
	var err error
	defer func() {
		// 统一的异常处理
		cb.ErrorProcess(c, "CheckTmdbApiHandler", err)
	}()

	req := tmdb_api.Req{}
	err = c.ShouldBindJSON(&req)
	if err != nil {
		return
	}
	if req.ApiKey == "" {
		c.JSON(http.StatusOK, backend.ReplyCommon{Message: "false"})
		return
	}
	// 备份一份
	bkProxySettings := settings.Get().AdvancedSettings.ProxySettings.CopyOne()
	// 赋值 Web 传递过来的需要测试的代理参数
	settings.Get().AdvancedSettings.ProxySettings = &req.ProxySettings
	// 设置代理
	err = local_http_proxy_server.SetProxyInfo(settings.Get().AdvancedSettings.ProxySettings.GetInfos())
	if err != nil {
		return
	}
	// 开始测试 tmdb api
	tmdbApi, err := tmdb_api.NewTmdbHelper(
		cb.fileDownloader.Log,
		req.ApiKey)
	if err != nil {
		cb.fileDownloader.Log.Errorln("NewTmdbHelper", err)
		return
	}

	defer func() {
		// 还原
		settings.Get().AdvancedSettings.ProxySettings = bkProxySettings
		cb.proxyCheckLocker.Unlock()
		err = local_http_proxy_server.SetProxyInfo(settings.Get().AdvancedSettings.ProxySettings.GetInfos())
		if err != nil {
			return
		}
		// 启动代理
		local_http_proxy_server.GetProxyUrl()
	}()

	if tmdbApi.Alive() == false {
		cb.fileDownloader.Log.Errorln("tmdbApi.Alive() == false")
		c.JSON(http.StatusOK, backend.ReplyCommon{Message: "false"})
		return
	} else {
		cb.fileDownloader.Log.Infoln("tmdbApi.Alive() == true")
		c.JSON(http.StatusOK, backend.ReplyCommon{Message: "true"})
		return
	}
}