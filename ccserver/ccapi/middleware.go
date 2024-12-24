package ccapi

import (
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
)

// ServerHTTPErrorHandler 自定义错误处理方法
//  错误信息直接返回一个只包含http状态码的响应
//  echo默认的错误处理会返回一个标准的RESTful API 一个json格式的错误响应
//  如果有大量的扫描服务器会影响服务器性能,将其替换为返回一个只包含http状态码的响应
//  HTTP/1.1 404 Not Found
//  Date: Thu, 23 Sep 2021 07:31:37 GMT
//  Content-Length: 0
func ServerHTTPErrorHandler(err error, c echo.Context) {
	he, ok := err.(*echo.HTTPError)
	if ok {
		if he.Internal != nil {
			if herr, ok := he.Internal.(*echo.HTTPError); ok {
				he = herr
			}
		}
	} else {
		he = &echo.HTTPError{
			Code:    http.StatusInternalServerError,
			Message: http.StatusText(http.StatusInternalServerError),
		}
	}

	code := he.Code
	message := he.Message

	// Send response
	if !c.Response().Committed {
		//err = c.String(code, "")
		log.Println(c.RealIP(), c.Request().RequestURI, code, message)
		err = c.NoContent(he.Code)
	}
}
