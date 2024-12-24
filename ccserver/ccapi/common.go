package ccapi

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

type H map[string]interface{}

func Fail(c echo.Context, code int, message string) error {
	return c.JSON(200, H{
		"code":    code,
		"message": message,
	})
}

func FailWithData(c echo.Context, code int, message string, data interface{}) error {
	return c.JSON(200, H{
		"code":    code,
		"message": message,
		"data":    data,
	})
}

func Success(c echo.Context, data interface{}) error {
	return c.JSON(200, H{
		"code":    1,
		"message": "success",
		"data":    data,
	})
}

func NotFound(c echo.Context, message string) error {
	return c.JSON(200, H{
		"code":    -1,
		"message": message,
	})
}

func HttpStatus200(c echo.Context, msg string) error {
	return c.String(http.StatusOK, msg)
}

func HttpStatus500(c echo.Context, msg string) error {
	return c.String(http.StatusInternalServerError, msg)
}

// HttpStatusByte200 返回二进制数据流
func HttpStatusByte200(c echo.Context, byte []byte) error {
	return c.Blob(http.StatusOK, echo.MIMEOctetStream, byte)
}
