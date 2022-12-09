package routers

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

/**
 * @api {get} /restart 重启服务
 * @apiGroup sys
 * @apiName Restart
 * @apiUse simpleSuccess
 */
func (h *APIHandler) Restart(c *gin.Context) {
	log.Println("Restart...")
	c.JSON(http.StatusOK, "OK")
	go func() {
		select {
		case h.RestartChan <- true:
		default:
		}
	}()
}
