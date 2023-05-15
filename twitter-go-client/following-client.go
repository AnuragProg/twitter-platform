package main

import (
	"fmt"
	"io"
	"log"
	"sync"
	"net/http"
)

func BulkFollowSpecificPerson(audience []*SignUpResponse, followeeId string){
	var wg sync.WaitGroup
	
	for _, person := range audience{
		if person != nil{
			wg.Add(1)	
			go func(person *SignUpResponse){
				Follow(person.Data.Token, followeeId)	
				wg.Done()
			}(person)
		}
	}
	
	wg.Wait()
}

func BulkFollow(jwtToken string, followees []string, bulkCount int){

	var wg sync.WaitGroup
	buff := make(chan int, bulkCount)

	for i, followeeId := range followees{

		log.Printf("%v. Request", i)
		wg.Add(1)
		buff<-i

		go func(followeeId string){
			Follow(jwtToken, followeeId)
			wg.Done()
			<-buff
		}(followeeId)
	}

	wg.Wait()
}


func Follow(jwtToken string, followeeId string){
	req, err := http.NewRequest("POST", fmt.Sprintf("%v/user/follow/%v", BASE_URL, followeeId), nil)
	if err != nil{
		log.Println(err.Error())
		return
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", jwtToken))
	res, err := http.DefaultClient.Do(req)
	if err != nil{
		log.Println(err.Error())
		return
	}
	defer res.Body.Close()
	
	resBody, err := io.ReadAll(res.Body)
	if err != nil{
		log.Println(err.Error())
		return
	}
	log.Println(string(resBody))
}