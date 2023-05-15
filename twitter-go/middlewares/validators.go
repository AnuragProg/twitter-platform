package middlewares

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	util "twitter-go/utils"

	"github.com/gin-gonic/gin"
)


func MobileValidator(ctx *gin.Context){
	var data struct{
		Mobile string `json:"mobile"`
	}
	
	bodyBytes, err := io.ReadAll(ctx.Request.Body)	
	ctx.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if err!=nil{
		util.HandleError(ctx, http.StatusBadRequest, errors.New("mobile missing"))	
		ctx.Abort()
		return
	}
	
	if err = json.Unmarshal(bodyBytes, &data); err!=nil{
		util.HandleError(ctx, http.StatusBadRequest, errors.New("mobile missing"))	
		ctx.Abort()
		return
	}

	pattern := regexp.MustCompile("^[0-9]{10}$")	
	
	if !pattern.Match([]byte(data.Mobile)){
		util.HandleError(ctx, http.StatusBadRequest, errors.New("invalid mobile number"))
		ctx.Abort()
		return
	}
}