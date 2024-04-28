package utils

import (
	"log"

	"github.com/gin-gonic/gin"
)

type Gin struct {
	C *gin.Context
}

type Response struct {
	Code int    `json:"code" example:"200"`
	Msg  string `json:"msg" example:"ok"`
	Data any    `json:"data" example:""`
	URI  string `json:"uri,omitempty" example:"/api/v1/"`
}

type ErrorWithDetails struct {
	Error   string `json:"error"`
	Details any    `json:"details"`
}

// Response setting gin.JSON
func (g *Gin) Response(httpCode, errCode int, data any) {
	g.C.JSON(httpCode, Response{
		Code: errCode,
		Msg:  GetMsg(errCode),
		Data: data,
		URI:  g.C.Request.RequestURI,
	})
}

func (g *Gin) ResponseWithLog(httpCode, errCode int, data any) {
	defer log.Printf("api %s response: %+v", g.C.Request.RequestURI, data)
	g.Response(httpCode, errCode, data)
}

const (
	SUCCESS        = 200
	ERROR          = 500
	INVALID_PARAMS = 400
)

var MsgFlags = map[int]string{
	SUCCESS:        "ok",
	ERROR:          "fail",
	INVALID_PARAMS: "请求参数错误",
}

func GetMsg(code int) string {
	msg, ok := MsgFlags[code]
	if ok {
		return msg
	}

	return MsgFlags[ERROR]
}
