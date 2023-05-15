package utils

import (
	"strings"
	"github.com/gin-gonic/gin"
)

func formatSuccessMsg(msg *string){
	if len(*msg) > 1{
		*msg = strings.Trim(strings.ToUpper(string((*msg)[0])) + (*msg)[1:], " ") + "."
	}else if len(*msg) == 1{
		*msg = strings.ToUpper(*msg) + "."	
	}
}

func HandleSuccess(ctx *gin.Context, code int, msg string){
	formatSuccessMsg(&msg)
	ctx.JSON(code, gin.H{"success":true, "msg": msg})
}


func HandleSuccessWithData(ctx *gin.Context, code int, data interface{}){
	ctx.JSON(code, gin.H{"success": true, "data": data})
}