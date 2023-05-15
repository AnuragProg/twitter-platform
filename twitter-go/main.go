package main

import (
	"os"
	"fmt"
	ctrl "twitter-go/controllers"
	mid "twitter-go/middlewares"
	"twitter-go/models"

	// "github.com/google/uuid"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)


type Server struct{
	RDB *redis.Client
	DB *gorm.DB
}


func (s *Server) Setup(){
	
	if err := godotenv.Load(); err!=nil{
		panic("unable to load .env file")
	}

	RDB := redis.NewClient(&redis.Options{
        Addr:     "redis:6379",
        Password: "", // no password set
        DB:       0,  // use default DB
    })
	
	dsn := "host=postgres user=root password=password dbname=twitter port=5432 sslmode=disable TimeZone=Asia/Kolkata"
	DB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil{
		panic(err)
	}
	

	s.RDB = RDB
	s.DB = DB

	DB.Migrator().DropTable(&models.User{}, &models.Feed{}, &models.Following{})
	DB.AutoMigrate(&models.User{}, &models.Feed{}, &models.Following{})
}

func (s *Server) TearDown() {

	s.RDB.Close()	

	DBPostGres, err := s.DB.DB()
	if err != nil{
		panic(err)
	}
	
	DBPostGres.Close()
}

func setupRouter(s *Server) *gin.Engine{
	app := gin.Default()
	
	userHandler := &ctrl.UserHandler{
		DB: s.DB,
	}
	authHandler := &mid.AuthHandler{
		DB: s.DB,
	}
	
	feedHandler := &ctrl.FeedHandler{
		DB: s.DB,
		RDB: s.RDB,
	}
	
	v1 := app.Group("/api/v1")
	
	// acronym for User Group
	ug := v1.Group("/user")
	{
		ug.POST("/signUp", mid.MobileValidator, userHandler.SignUp)	
		ug.POST("/signIn", mid.MobileValidator, userHandler.SignIn)
		ug.POST("/follow/:id", authHandler.UserAuth, userHandler.Follow)
	}
	
	// acronym for Feed Group
	fg := v1.Group("/feed")
	{
		fg.POST("/post", authHandler.UserAuth, feedHandler.Post)
		fg.GET("/dashboard/db", authHandler.UserAuth, feedHandler.DashboardFromDB)
		fg.GET("/dashboard", authHandler.UserAuth, feedHandler.DashboardFromCacheAndDB)
	}
	
	return app
}

func main(){
	server := &Server{}	
	server.Setup()
	defer server.TearDown()
	
	app := setupRouter(server)

	app.Run(fmt.Sprintf(":%v",os.Getenv("PORT")))
}
