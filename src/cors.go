package src

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

var Cors = func() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var origins []string
		var headers []string
		var methods []string
		if ctx.Request.Method == http.MethodOptions {
			// 当跨域请求包含 header 数据时, 浏览器会首先发送一次 OPTIONS 请求, 请求中包含了跨域相关的信息,
			// 服务器经过判断确定请求合法时, 再返回相关的允许跨域的信息, 浏览器通过跨域验证后, 才会真正的发送跨域请求
			origins = ctx.Request.Header[headerKeyRequestOrigin]
			headers = ctx.Request.Header[headerKeyRequestHeaders]
			methods = ctx.Request.Header[headerKeyRequestMethod]
		} else {
			origin := ctx.Request.Header.Get("Origin")
			if origin == "" {
				origin = "*"
			}
			origins = []string{origin}
			headers = DefaultCorsHeaders
			methods = []string{ctx.Request.Method}
		}
		writerHeader := ctx.Writer.Header()
		AllowCrossSiteOrigin(writerHeader, origins)
		AllowCrossSiteHeaders(writerHeader, headers)
		AllowCrossSiteMethods(writerHeader, methods)
		ExposeCrossSiteHeaders(writerHeader, DefaultExposeHeaders)
	}
}()

// 服务器响应给客户端的跨域头信息
var (
	headerKeyAccessHeaders     = "Access-Control-Allow-Headers"
	headerKeyAccessMethods     = "Access-Control-Allow-Methods"
	headerKeyAccessOrigin      = "Access-Control-Allow-Origin"
	headerKeyAccessCredentials = "Access-Control-Allow-Credentials"
	exposeHeaders              = "Access-Control-Expose-Headers"
)

// 客户端请求服务器的跨域头信息
var (
	headerKeyRequestOrigin  = "Origin"
	headerKeyRequestHeaders = "Access-Control-Request-Headers"
	headerKeyRequestMethod  = "Access-Control-Request-Method"
)

// 允许跨域请求头
func AllowCrossSiteHeaders(header http.Header, headers []string) {
	header[headerKeyAccessHeaders] = append(header[headerKeyAccessHeaders], headers...)
}

// 允许跨域请求方法
func AllowCrossSiteMethods(header http.Header, methods []string) {
	header[headerKeyAccessMethods] = append(header[headerKeyAccessMethods], methods...)
}

// 允许跨域请求站点
func AllowCrossSiteOrigin(header http.Header, origin []string) {
	header[headerKeyAccessOrigin] = append(header[headerKeyAccessOrigin], origin...)
}

// 允许跨域 cookies, 如果开启, 则 Access-Control-Allow-Origin 响应头必须为具体网站地址, 不可为 *
func AllowCrossSiteCredentials(header http.Header) {
	header.Set(headerKeyAccessCredentials, "true")
}

// 允许浏览器操作的响应头
func ExposeCrossSiteHeaders(header http.Header, headers []string) {
	header[exposeHeaders] = append(header[exposeHeaders], headers...)
}

// 默认允许跨域请求的 header 头
var DefaultCorsHeaders = []string{"Content-Type", "Authorization"}

// 默认允许跨域响应的 header 头
var DefaultExposeHeaders = []string{"Middleware"}