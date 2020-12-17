package handler

import (
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"strings"
)

type Proxy struct {
}

// 图片代理
func (p *Proxy) ProxyImage(ctx *gin.Context) {
	url := ctx.Query("img")
	if url != "" {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			SendError(ctx, err)
		} else {
			for key, vs := range ctx.Request.Header {
				if strings.ToLower(key) != "referer" {
					req.Header[key] = vs
				}
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				SendError(ctx, err)
			} else {
				defer resp.Body.Close()
				data, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					SendError(ctx, err)
				} else {
					h := ctx.Writer.Header()
					for k, vs := range resp.Header {
						h[k] = vs
					}
					ctx.Writer.Write(data)
				}
			}
		}
	}
}
