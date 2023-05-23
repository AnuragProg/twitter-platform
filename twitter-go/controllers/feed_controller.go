package controllers

import (
	"context"
	"errors"
	"log"

	// "math/rand"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	model "twitter-go/models"
	util "twitter-go/utils"
)


type FeedHandler struct{
	DB *gorm.DB	
	RDB *redis.Client
}


func (f *FeedHandler) Post(ctx *gin.Context){
	var feedRequest struct{
		Content string `json:"content"`
	}	
	
	if err:=ctx.BindJSON(&feedRequest); err != nil{
		util.HandleError(ctx, http.StatusBadRequest, err)
		return
	}
	
	userId, exists := ctx.Get("userId")
	if !exists{
		util.HandleError(ctx, http.StatusInternalServerError, errors.New("userid not present in access token"))
		return
	}
	parsedUserId, err := uuid.Parse(userId.(string))
	if err != nil{
		util.HandleError(ctx, http.StatusInternalServerError, err)
		return
	}	
	log.Println("Posted User:", parsedUserId.String())
	newFeed := &model.Feed{
		ID: uuid.New(),
		UserId: parsedUserId,
		Content: feedRequest.Content,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := f.DB.Create(newFeed).Error; err != nil{
		util.HandleError(ctx, http.StatusInternalServerError, err)
		return
	}
	
	// Pushing to cache of users
	go f.pushFeedToCache(newFeed, newFeed.UserId.String())

	util.HandleSuccess(ctx, http.StatusOK, "feed posted successfully")
}

func (f *FeedHandler)pushFeedToCache(feed *model.Feed, ownerId string){
	defer func(){
		if err:=recover(); err!=nil{
			log.Println(err)
		}
	}()

	// Getting follower counts
	var followersCount int64
	if err := f.DB.Model(&model.Following{}).Where(&model.Following{FolloweeId: uuid.MustParse(ownerId)}).Count(&followersCount).Error; err!=nil{
		log.Println(err.Error())
		return
	}

	log.Println("Followers Count:", followersCount)
	// Will ba directly querying from db for people who have more than FOLLOWER_COUNT_THRESHOLD followers	
	if followersCount > util.FOLLOWER_COUNT_THRESHOLD{
		log.Println("Not pushing post to cache")
		return	
	}

	
	feedJson, _ := json.Marshal(feed)
	
	var wg sync.WaitGroup
	var followerIds []string

	f.DB.Model(&model.Following{}).Where("followee_id = ?", uuid.MustParse(ownerId)).Pluck("follower_id", &followerIds)
	buf := make(chan int, 8)

	for _, follower := range followerIds{
		buf<-0
		wg.Add(1)
		go func(id string){
			defer wg.Done()
			log.Printf("Cache Push : UserId: %v | Feed: %v", id, feedJson)
			f.RDB.RPush(context.TODO(), id, feedJson)
			// Setting expire time to 48 hours if feed is already present
			if exists, _ := f.RDB.Exists(context.TODO(), id).Result(); exists == 1{
				f.RDB.Expire(context.TODO(), id, 48*time.Hour)				
			}
			<-buf
		}(follower)
	}
	wg.Wait()
}


func (f *FeedHandler) DashboardFromCacheAndDB(ctx *gin.Context){
	var feedsFromDB []*model.Feed
	var feedsFromCache []*model.Feed

	
	currentContext, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	var feedsJSONFromCache []string
	err := f.RDB.LRange(currentContext, ctx.GetString("userId"), 0, -1).ScanSlice(&feedsJSONFromCache)
	if err != nil{
		util.HandleError(ctx, http.StatusInternalServerError, err)
		return
	}
	
	for _, feedJSON := range feedsJSONFromCache{
		var feed model.Feed
		json.Unmarshal([]byte(feedJSON), &feed)
		feedsFromCache = append(feedsFromCache, &feed)
	}

	log.Printf("Cache Feeds = %v", feedsFromCache)
	
	// err = f.DB.Raw("SELECT feeds.* FROM feeds INNER JOIN followings ON followings.followee_id = feeds.user_id WHERE followings.follower_id = ? AND (SELECT COUNT(*) FROM followings WHERE followee_id = followings.followee_id) > ?", uuid.MustParse(ctx.GetString("userId")), util.FOLLOWER_COUNT_THRESHOLD).Scan(&feedsFromDB).Error
	// err = f.DB.Raw(` 	
	// 	SELECT feeds.* 
	// 	FROM followings
	// 	JOIN feeds ON feeds.user_id = followings.followee_id
	// 	JOIN (
	// 	  SELECT followee_id, COUNT(*) AS num_followers
	// 	  FROM followings
	// 	  GROUP BY followee_id
	// 	  HAVING COUNT(*) > ?
	// 	) AS popular_users ON popular_users.followee_id = feeds.user_id
	// 	WHERE followings.follower_id = ?
	// `, util.FOLLOWER_COUNT_THRESHOLD, uuid.MustParse(ctx.GetString("userId"))).Scan(&feedsFromDB).Error
	
	err = f.DB.Raw(`
		SELECT * from feeds 
		WHERE user_id IN (
			SELECT followee_id from followings
			WHERE followee_id IN ( SELECT followee_id from followings WHERE follower_id = ? )
			GROUP BY followee_id
			HAVING COUNT(*) > ?
		)
	`, uuid.MustParse(ctx.GetString("userId")), util.FOLLOWER_COUNT_THRESHOLD).Scan(&feedsFromDB).Error
	if err != nil{
		util.HandleError(ctx, http.StatusInternalServerError, err)
		return
	}
	
	log.Printf("DB Feeds = %v", feedsFromDB)

	// r := rand.New(rand.NewSource(time.Now().UnixMilli()))
	// if len(feedsFromCache) != 0{
	// 	r.Shuffle(len(feedsFromCache), func (i, j int)  {
	// 		temp := feedsFromCache[i]	
	// 		feedsFromCache[i] = feedsFromCache[j]
	// 		feedsFromCache[j] = temp
	// 	})
	// }
	// if len(feedsFromDB) != 0{
	// 	r.Shuffle(len(feedsFromDB), func (i, j int)  {
	// 		temp := feedsFromDB[i]	
	// 		feedsFromDB[i] = feedsFromDB[j]
	// 		feedsFromDB[j] = temp
	// 	})
	// }

	finalFeedsList := append(feedsFromCache, feedsFromDB...)
	log.Printf("Final Feeds List = %v", finalFeedsList)
	util.HandleSuccessWithData(ctx, http.StatusOK, finalFeedsList)
}


func (f *FeedHandler) DashboardFromDB(ctx *gin.Context){
	var feeds []*model.Feed	
	
	parsedUserId, err := uuid.Parse(ctx.GetString("userId"))
	if err != nil{
		util.HandleError(ctx, http.StatusUnauthorized, errors.New("invalid bearer token"))
		return
	}
	
	if err:=f.DB.Raw("SELECT feeds.* FROM feeds INNER JOIN followings ON followings.followee_id = feeds.user_id WHERE followings.follower_id = ?", parsedUserId).Scan(&feeds).Error; err!=nil{
		util.HandleError(ctx, http.StatusInternalServerError, err)
		return
	}
	
	util.HandleSuccessWithData(ctx, http.StatusOK, feeds)	
}