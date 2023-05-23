package main

import (
	"log"
	"time"
	"math"
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	BASE_URL    string = "http://0.0.0.0:3000/api/v1"
	NO_OF_USERS uint   = 1_000
)


type Benchmark struct{
	gorm.Model
	Timing int64 // Unix time millis
	Type uint // 1 - From DB only ; 2 - From DB & Cache
	NoOfItemsInFeed int `gorm:"column:no_of_items_in_feed"`
}

type Storage struct{
	DB *gorm.DB	
}

// Write doc for the function
func main() {
	
	db, err := gorm.Open(sqlite.Open("benchmark.db"), &gorm.Config{})
	if err != nil{
		log.Fatal(err.Error())
	}
	
	db.Migrator().DropTable(&Benchmark{})
	db.AutoMigrate(&Benchmark{})
	
	s := &Storage{DB: db}
	bc := &BenchmarkConfig{
		NormalInfluencerCount: 10,
		FamousInfluencerCount: 10,
		AudienceCount: 400,
		BufferAudienceCount: 100,

		BatchSize: 50,	

		NormalInfluencerPostCount: 50,
		FamousInfluencerPostCount: 50,

		AudienceMutex: &sync.Mutex{},
		BufferAudienceMutex: &sync.Mutex{},
		NormalInfluencerMutex: &sync.Mutex{},
		FamousInfluencerMutex: &sync.Mutex{},
		
		Audience: []*SignUpResponse{},
		BufferAudience: []*SignUpResponse{},
		NormalInfluencer: []*SignUpResponse{},
		FamousInfluencer: []*SignUpResponse{},
	}
	bc.BenchmarkDashboard(s)
	
	// For testing whether backend is working properly

	// influencer := SignUp(GenerateFakeSignUpRequest())
	// normalInfluencer := SignUp(GenerateFakeSignUpRequest())
	
	// audience1 := SignUp(GenerateFakeSignUpRequest())
	// audience2 := SignUp(GenerateFakeSignUpRequest())
	// audience3 := SignUp(GenerateFakeSignUpRequest())
	
	// Follow(audience1.Data.Token, influencer.Data.UserId)
	// Follow(audience2.Data.Token, influencer.Data.UserId)
	// Follow(audience3.Data.Token, influencer.Data.UserId)
	// Follow(audience1.Data.Token, normalInfluencer.Data.UserId)
	
	// BulkPostFeed(influencer.Data.Token, 5, 5)
	// BulkPostFeed(normalInfluencer.Data.Token, 5, 5)
	
	// dashboard := GetDashboard(audience1.Data.Token)
	// for _, feed := range dashboard{
	// 	log.Printf("%v - %v",feed.ID, feed.Content[:20])
	// }
}



type BenchmarkConfig struct{
	NormalInfluencerCount int
	FamousInfluencerCount int	
	AudienceCount int
	BufferAudienceCount int
	
	NormalInfluencerPostCount int
	FamousInfluencerPostCount int	

	BatchSize int
	
	Audience []*SignUpResponse
	BufferAudience []*SignUpResponse
	NormalInfluencer []*SignUpResponse
	FamousInfluencer []*SignUpResponse
	
	AudienceMutex *sync.Mutex
	BufferAudienceMutex *sync.Mutex
	NormalInfluencerMutex *sync.Mutex
	FamousInfluencerMutex *sync.Mutex
}
type Type int

const (
	AUDIENCE = iota
	BUFFER_AUDIENCE
	NORMAL_INFLUENCER
	FAMOUS_INFLUENCER
)

func (b *BenchmarkConfig)PrintCountOfUsers(){
	log.Printf("Normal Influencer Count: %v", len(b.NormalInfluencer))
	log.Printf("Famous Influencer Count: %v", len(b.FamousInfluencer))
	log.Printf("Audience Count: %v", len(b.Audience))
	log.Printf("Buffer Audience Count: %v", len(b.BufferAudience))
}

func (b *BenchmarkConfig)Append(val *SignUpResponse, t Type){
	if val == nil{
		return
	}
	switch t{
		case AUDIENCE:
			b.AudienceMutex.Lock()
			b.Audience = append(b.Audience, val)
			b.AudienceMutex.Unlock()
		case BUFFER_AUDIENCE:
			b.BufferAudienceMutex.Lock()
			b.BufferAudience = append(b.BufferAudience, val)
			b.BufferAudienceMutex.Unlock()
		case NORMAL_INFLUENCER:
			b.NormalInfluencerMutex.Lock()
			b.NormalInfluencer = append(b.NormalInfluencer, val)
			b.NormalInfluencerMutex.Unlock()
		case FAMOUS_INFLUENCER:
			b.FamousInfluencerMutex.Lock()
			b.FamousInfluencer = append(b.FamousInfluencer, val)	
			b.FamousInfluencerMutex.Unlock()
		default :
			log.Fatal("Invalid type")
	}
}

type Number interface{
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64
}

func min[A Number](a, b A) A{
	if a < b{
		return a
	}
	return b
}


func (b *BenchmarkConfig)BenchmarkDashboard(s *Storage){
	
	var wg sync.WaitGroup
	
	// Signing Up Normal Influencers
	for i:=0; i<b.NormalInfluencerCount; i+=b.BatchSize{
		currentBatch := min(b.BatchSize, b.NormalInfluencerCount-i)
		for j:=0; j<currentBatch; j++{
			wg.Add(1)
			go func(){
				defer wg.Done()
				b.Append(SignUp(GenerateFakeSignUpRequest()), NORMAL_INFLUENCER)
			}()
		}
		wg.Wait()
	}

	// Signing Up Famous Influencers
	for i:=0; i<b.FamousInfluencerCount; i+=b.BatchSize{
		currentBatch := min(b.BatchSize, b.FamousInfluencerCount-i)
		for j:=0; j<currentBatch; j++{
			wg.Add(1)
			go func(){
				defer wg.Done()
				b.Append(SignUp(GenerateFakeSignUpRequest()), FAMOUS_INFLUENCER)
			}()
		}		
		wg.Wait()
	}

	// Signing Up Audience(that will follow both normal influencers and famous influencers)
	for i:=0; i<b.AudienceCount; i+=b.BatchSize{
		currentBatch := min(b.BatchSize, b.AudienceCount-i)
		for j:=0; j<currentBatch; j++{
			wg.Add(1)
			go func(){
				defer wg.Done()
				b.Append(SignUp(GenerateFakeSignUpRequest()), AUDIENCE)
			}()
		}		
		wg.Wait()
	}

	// Signing Up Buffer Audience(that will follow only famous influencers)
	for i:=0; i<b.BufferAudienceCount; i+=b.BatchSize{
		currentBatch := min(b.BatchSize, b.BufferAudienceCount-i)
		for j:=0; j<currentBatch; j++{
			wg.Add(1)
			go func(){
				defer wg.Done()
				b.Append(SignUp(GenerateFakeSignUpRequest()), BUFFER_AUDIENCE)
			}()
		}		
		wg.Wait()
	}

	// Printing Count of Users to see how many users were created
	b.PrintCountOfUsers()

	// Audience Following Normal Influencers & Famous Influencers
	totalBatches := math.Ceil(float64(len(b.Audience))/ float64(b.BatchSize))
	tempAudience := make([]*SignUpResponse, len(b.Audience))
	copy(tempAudience, b.Audience)
	for i:=0; i<int(totalBatches); i++{
		currentBatch := tempAudience[:min(b.BatchSize, len(tempAudience))]
		tempAudience = tempAudience[min(b.BatchSize, len(tempAudience)):]
		for _, audience := range currentBatch{
			if(audience == nil) {
				continue
			}
			wg.Add(1)
			// Following Normal Influencers
			go func(audience *SignUpResponse){
				defer wg.Done()
				for _, influencer := range b.NormalInfluencer{
					Follow(audience.Data.Token, influencer.Data.UserId)
				}
				for _, influencer := range b.FamousInfluencer{
					Follow(audience.Data.Token, influencer.Data.UserId)
				}
			}(audience)
		}
		wg.Wait()
	}
	
	// Buffer Audience Following Normal Influencers & Famous Influencers
	totalBatches = math.Ceil(float64(len(b.BufferAudience))/ float64(b.BatchSize))
	tempBufferAudience := make([]*SignUpResponse, len(b.BufferAudience))
	copy(tempBufferAudience, b.BufferAudience)
	for i:=0; i<int(totalBatches); i++{
		currentBatch := tempBufferAudience[:min(b.BatchSize, len(tempBufferAudience))]
		tempBufferAudience = tempBufferAudience[min(b.BatchSize, len(tempBufferAudience)):]
		for _, audience := range currentBatch{
			if(audience == nil) {
				continue
			}
			wg.Add(1)
			
			// Following Normal Influencers
			go func(audience *SignUpResponse){
				defer wg.Done()
				for _, influencer := range b.FamousInfluencer{
					Follow(audience.Data.Token, influencer.Data.UserId)
				}
			}(audience)
		}
		wg.Wait()
	}
	
	
	// Posting Feeds by Normal Influencers
	totalBatches = math.Ceil(float64(len(b.NormalInfluencer))/ float64(b.BatchSize))
	tempNormalInfluencer := make([]*SignUpResponse, len(b.NormalInfluencer))
	copy(tempNormalInfluencer, b.NormalInfluencer)
	for i:=0; i<int(totalBatches); i++{
		currentBatch := tempNormalInfluencer[:min(b.BatchSize, len(tempNormalInfluencer))]
		tempNormalInfluencer = tempNormalInfluencer[min(b.BatchSize, len(tempNormalInfluencer)):]
		for _, influencer := range currentBatch{
			if influencer == nil{
				continue
			}
			wg.Add(1)
			go func(influencer *SignUpResponse){
				defer wg.Done()
				PostFeed(influencer.Data.Token, GenerateFakeFeedRequest())
			}(influencer)
		}
		wg.Wait()
	}

	// Posting Feeds by Famous Influencers
	totalBatches = math.Ceil(float64(len(b.FamousInfluencer))/ float64(b.BatchSize))
	tempFamousInfluencer := make([]*SignUpResponse, len(b.FamousInfluencer))
	copy(tempFamousInfluencer, b.FamousInfluencer)
	for i:=0; i<int(totalBatches); i++{
		currentBatch := tempFamousInfluencer[:min(b.BatchSize, len(tempFamousInfluencer))]
		tempFamousInfluencer = tempFamousInfluencer[min(b.BatchSize, len(tempFamousInfluencer)):]
		for _, influencer := range currentBatch{
			if influencer == nil{
				continue
			}
			wg.Add(1)
			go func(influencer *SignUpResponse){
				defer wg.Done()
				PostFeed(influencer.Data.Token, GenerateFakeFeedRequest())
			}(influencer)
		}
		wg.Wait()
	}
	
	time.Sleep(time.Second * 5)
	
	// Retrieving Dashboard of Audience From DB only and saving timings in db
	totalBatches = math.Ceil(float64(len(b.Audience)+len(b.BufferAudience))/ float64(b.BatchSize))
	tempAudience = make([]*SignUpResponse, len(b.Audience) + len(b.BufferAudience))
	copy(tempAudience, b.Audience)
	tempAudience = append(tempAudience, b.BufferAudience...)
	for i:=0; i<int(totalBatches); i++{
		currentBatch := tempAudience[:min(b.BatchSize, len(tempAudience))]
		tempAudience = tempAudience[min(b.BatchSize, len(tempAudience)):]
		for _, audience := range currentBatch{
			if audience == nil{
				continue
			}
			wg.Add(1)
			go func(audience *SignUpResponse){
				defer wg.Done()
				now := time.Now()
				noOfItemsInFeed := len(GetDashboardFromDBOnly(audience.Data.Token))
				s.DB.Create(&Benchmark{Timing: int64(time.Since(now).Milliseconds()), Type: 1, NoOfItemsInFeed: noOfItemsInFeed})
			}(audience)
		}
		wg.Wait()
	}
	
	time.Sleep(time.Second * 5)
	
	// Retrieving Dashboard of Audience From DB And Cache and saving timings in db
	totalBatches = math.Ceil(float64(len(b.Audience))/ float64(b.BatchSize))
	tempAudience = make([]*SignUpResponse, len(b.Audience))
	copy(tempAudience, b.Audience)
	for i:=0; i<int(totalBatches); i++{
		currentBatch := tempAudience[:min(b.BatchSize, len(tempAudience))]
		tempAudience = tempAudience[min(b.BatchSize, len(tempAudience)):]
		for _, audience := range currentBatch{
			if audience == nil{
				continue
			}
			wg.Add(1)
			go func(audience *SignUpResponse){
				defer wg.Done()
				now := time.Now()
				noOfItemsInFeed := len(GetDashboard(audience.Data.Token))
				s.DB.Create(&Benchmark{Timing: int64(time.Since(now).Milliseconds()), Type: 2, NoOfItemsInFeed: noOfItemsInFeed})
			}(audience)
		}
		wg.Wait()
	}
	
}

