package middleware

import (
	"api-gateway/pkg/e"
	"api-gateway/proxy"
	"encoding/json"
	"fmt"
	mapset "github.com/deckarep/golang-set"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
)

const (
	_userServerKey = "/v1/user/http"
	_userUrl = "http://%s/api/v1/user/auth-check"
)

// todo: 无需鉴权的接口可设置公共平台配置(如果有)
var freePaths = mapset.NewSet(
	"/api/v1/user/register",
	"/api/v1/user/login",
)

// JWT token验证中间件
func JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		var code int
		var data interface{}
		code = 200

		if freePaths.Contains(c.Request.URL.Path) {
			c.Next()
			return
		}

		key := "Authorization"
		token := c.GetHeader(key)
		if token == "" {
			code = 401
		} else {
			// 采用 user 模块的接口鉴权
			userServer := proxy.GetServer(_userServerKey)
			url := fmt.Sprintf(_userUrl, userServer.Addr)

			req, err := http.NewRequest(http.MethodPost, url, nil)
			if err != nil {
				code = e.ErrorAuthCheckTokenFail
			} else {
				req.Header.Set(key, token)
				res, err := (&http.Client{}).Do(req)
				if err != nil || res == nil || res.StatusCode != http.StatusOK || res.Body == nil {
					code = e.ErrorAuthCheckTokenFail
				} else {
					resCode := &struct {
						Code int `json:"code"`
					}{}
					body, _ := ioutil.ReadAll(res.Body)
					err = json.Unmarshal(body, resCode)
					if err != nil || resCode.Code != 200 {
						code = e.ErrorAuthCheckTokenFail
					}
				}
			}
		}
		if code != e.SUCCESS {
			c.JSON(200, gin.H{
				"status": code,
				"msg":    e.GetMsg(uint(code)),
				"data":   data,
			})
			c.Abort()
			return
		}
		c.Next()
		return
	}
}