package utils

import (
	"log"
	"strings"

	"github.com/gin-gonic/gin"
)

func formatErrorMsg(msg *string){
	if len(*msg) > 1{
		*msg = strings.Trim(strings.ToUpper(string((*msg)[0])) + (*msg)[1:], " ") + "!"
	}else if len(*msg) == 1{
		*msg = strings.ToUpper(*msg) + "!"	
	}
}

func HandleError(ctx *gin.Context, code int, err error){
	errMsg := err.Error()
	formatErrorMsg(&errMsg)
	log.Printf("Formatted Error Msg: %v", errMsg)	
	ctx.JSON(code, gin.H{"success": false, "msg": errMsg})
}