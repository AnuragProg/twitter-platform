package middlewares

import (
	"errors"
	"log"
	"net/http"
	"strings"
	util "twitter-go/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthHandler struct{
	DB *gorm.DB
}

func (a *AuthHandler)UserAuth(ctx *gin.Context){
	authHeader := strings.Split(ctx.Request.Header.Get("Authorization"), " ")
	if len(authHeader) != 2 || strings.Trim(authHeader[1], " ") == ""{
		log.Println("Got Authorization", authHeader)	
		util.HandleError(ctx, http.StatusUnauthorized, errors.New("invalid bearer token"))
		ctx.Abort()
		return
	}

	userId, _, err := util.VerifyToken(authHeader[1])
	if err != nil{
		util.HandleError(ctx, http.StatusUnauthorized, err)
		ctx.Abort()
		return
	}
	log.Println("Verified user id =", userId)
	ctx.Set("userId", userId)
}


func (a *AuthHandler)AdminAuth(ctx *gin.Context){
	authHeader := strings.Split(ctx.Request.Header.Get("Authorization"), " ")
	if len(authHeader) != 2 || strings.Trim(authHeader[1], " ") != ""{
		util.HandleError(ctx, http.StatusUnauthorized, errors.New("invalid bearer token"))
		return
	}

	userId, isAdmin, err := util.VerifyToken(authHeader[1])
	if err != nil{
		util.HandleError(ctx, http.StatusUnauthorized, err)
		ctx.Abort()
		return
	}
	if !isAdmin{
		util.HandleError(ctx, http.StatusUnauthorized, errors.New("unauthroized access"))	
		ctx.Abort()
		return
	}
	ctx.Set("userId", userId)
}