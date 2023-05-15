package controllers

import (
	"errors"
	"log"
	"net/http"
	"time"
	model "twitter-go/models"
	util "twitter-go/utils"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserHandler struct {
	DB *gorm.DB
}


func (u *UserHandler) SignUp(ctx *gin.Context){
	log.Println("Starting SignUp")
	var userRequest struct{
		Username string `json:"username"`
		Password string `json:"password"`
		Mobile string `json:"mobile"`
	}
	
	if err := ctx.ShouldBindWith(&userRequest, binding.JSON); err != nil{
		log.Println("Got error in Binding JSON error found")
		util.HandleError(ctx, http.StatusBadRequest, err)
		return
	}	
	if userRequest.Username == "" || userRequest.Password=="" || userRequest.Mobile==""{
		log.Println("Empty fields condition")
		util.HandleError(ctx, http.StatusBadRequest, errors.New("missing required fields"))
		return
	}
	
	
	if err := u.DB.Where(&model.User{Mobile: userRequest.Mobile}).First(&model.User{}).Error; err==nil{
		util.HandleError(ctx, http.StatusBadRequest, errors.New("user already exists"))
		return
	}
	// u.DB.FirstOrCreate(user)

	user := &model.User{
		ID: uuid.New(),
		Username: userRequest.Username,
		Password: userRequest.Password,
		Mobile: userRequest.Mobile,
		JoinedOn: time.Now(),
	}
	
	u.DB.Create(&user)
	
	tokenString, err := util.GenerateToken(user.Username, user.ID.String())
	
	if err!=nil{
		log.Println("Token error")
		util.HandleError(ctx, http.StatusInternalServerError, err)
		return	
	}


	util.HandleSuccessWithData(ctx, http.StatusOK, gin.H{
		"token": tokenString,
		"msg": "user signed up successfully",
		"userId": user.ID,
	})
}

func (u *UserHandler) SignIn(ctx *gin.Context){
	var userRequest struct{
		Mobile string `json:"mobile"`	
		Password string `json:"password"`
	}	
	
	if err := ctx.ShouldBindWith(&userRequest, binding.JSON); err!=nil{
		util.HandleError(ctx, http.StatusBadRequest, err)
		return
	}
	if userRequest.Password=="" || userRequest.Mobile==""{
		util.HandleError(ctx, http.StatusBadRequest, errors.New("missing required fields"))
		return
	}
	
	var user model.User
	if err := u.DB.Where(&model.User{Mobile:userRequest.Mobile, Password:userRequest.Password}).First(&user).Error; err!=nil{
		if errors.Is(err, gorm.ErrRecordNotFound){
			util.HandleError(ctx, http.StatusNotFound, errors.New("user not found"))
		}else{
			util.HandleError(ctx, http.StatusNotFound, err)
		}
		return
	}
	
	tokenString, _ := util.GenerateToken(user.Username, user.ID.String()) 
	
	util.HandleSuccessWithData(ctx, http.StatusOK, struct{
		UserId string `json:"userId"`
		Token string `json:"token"`
	 	Msg string `json:"msg"`
	}{
		UserId: user.ID.String(),
		Token: tokenString,
		Msg: "user signed in successfully",
	})
}


func (u *UserHandler)Follow(ctx *gin.Context){
	followeeId, found := ctx.Params.Get("id")
	if !found{
		util.HandleError(ctx, http.StatusBadRequest, errors.New("followee id not found"))
		return
	}	
	
	parsedFolloweeId, err := uuid.Parse(followeeId)
	
	if err != nil{
		util.HandleError(ctx, http.StatusBadRequest, err)
		return
	}
	
	if err = u.DB.Where(&model.User{ID: parsedFolloweeId}).First(&model.User{}).Error; err!=nil{
		util.HandleError(ctx, http.StatusNotFound, errors.New("followee not found"))
		return
	}

	followerId := ctx.GetString("userId")
	parsedFollowerId, err := uuid.Parse(followerId)
	if err!=nil{
		util.HandleError(ctx, http.StatusUnauthorized, err)
		return
	}
	
	if parsedFolloweeId.String() == parsedFollowerId.String(){
		util.HandleError(ctx, http.StatusUnauthorized, errors.New("not allowed"))
		return
	}
	
	following := &model.Following{
		ID: uuid.New(),
		FollowerId: parsedFollowerId,
		FolloweeId: parsedFolloweeId,
	}
	u.DB.Where(&model.Following{FollowerId: parsedFollowerId, FolloweeId: parsedFolloweeId}).FirstOrCreate(following)
	
	util.HandleSuccess(ctx, http.StatusOK, "successfully followed")
}