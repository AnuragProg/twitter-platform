package main

import (
	"io"
	"fmt"
	"log"
	"sync"
	"time"
	"bytes"
	"reflect"
	"net/http"
	"encoding/json"
	"github.com/go-faker/faker/v4"
)

type Feed struct {
	ID string `json:"id"`
	UserId string `json:"userId"`
	Content string `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type FeedRequest struct {
	Content string `json:"content"`	
}


func GenerateFakeFeedRequest()*FeedRequest{
	fakeContent, _ := faker.Lorem.Paragraph(faker.Lorem{}, reflect.Value{})
	return &FeedRequest{
		Content: fakeContent.(string),
	}
}

func GetDashboardFromDBOnly(jwtToken string)[]*Feed{
	req, err := http.NewRequest("GET", fmt.Sprintf("%v/feed/dashboard/db", BASE_URL), nil)	
	if err != nil{
		log.Println(err.Error())
		return nil
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", jwtToken))
	res, err := http.DefaultClient.Do(req)
	if err != nil{
		log.Fatal(err.Error())
	}
	defer res.Body.Close()

	bodyBytes, _ := io.ReadAll(res.Body)
	
	//log.Printf("Body JSON = %v",string(bodyBytes))
	var body struct{
		Data []*Feed `json:"data"`
		Success bool `json:"success"`
	}
	if err = json.Unmarshal(bodyBytes, &body); err !=nil{
		log.Println(err.Error())
		return nil
	}

	return body.Data

}

func GetDashboard(jwtToken string) []*Feed{
	req, err := http.NewRequest("GET", fmt.Sprintf("%v/feed/dashboard", BASE_URL), nil)	
	if err != nil{
		log.Println(err.Error())
		return nil
	}
	
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", jwtToken))
	res, err := http.DefaultClient.Do(req)
	if err != nil{
		log.Println(err.Error())		
		return nil	
	}
	defer res.Body.Close()

	bodyBytes, _ := io.ReadAll(res.Body)
	
	// log.Printf("Body JSON = %v",string(bodyBytes))
	var body struct{
		Data []*Feed `json:"data"`
		Success bool `json:"success"`
	}
	if err = json.Unmarshal(bodyBytes, &body); err !=nil{
		log.Println(err.Error())
		return nil
	}

	return body.Data
}

func PostFeed(jwtToken string, feedRequest *FeedRequest) map[string]interface{}{
	bodyBytes, _ := json.Marshal(feedRequest)
	body := bytes.NewBuffer(bodyBytes)
	
	log.Println("Request body:", string(bodyBytes))
	req, err := http.NewRequest("POST", fmt.Sprintf("%v/feed/post", BASE_URL), body)
	if err != nil{
		log.Println(err.Error())
		return nil
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", jwtToken))
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err !=nil{
		log.Println(err.Error())
		return nil
	}
	defer res.Body.Close()

	resBodyBytes, err := io.ReadAll(res.Body)
	if  err != nil{
		log.Println(err.Error())
		return nil
	}
	
	log.Println(string(resBodyBytes))
	
	var result map[string]interface{}
	json.Unmarshal(resBodyBytes, &result)
	return result	
}


func BulkPostFeed(jwtToken string, count, batchCount int){

	buffer := make(chan int, batchCount)
	var wg sync.WaitGroup
	for i:=0; i<count; i++{
		wg.Add(1)
		buffer<-i
		fmt.Printf("%v. Request", i+1)
		go func(){
			PostFeed(jwtToken, GenerateFakeFeedRequest())
			wg.Done()
			<-buffer
		}()
	}
	wg.Wait()
}