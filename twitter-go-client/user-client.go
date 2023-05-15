package main

import (
	"io"
	"log"
	"fmt"
	"sync"
	"bytes"
	"strings"
	"net/http"
	"encoding/json"
	"github.com/go-faker/faker/v4"
)

type SignUpRequest struct{
	Username string `json:"username"`	
	Password string `json:"password"`
	Mobile string `json:"mobile"`
}

type Data struct{
	UserId string `json:"userId"`
	Token string `json:"token"`	
	Msg string `json:"msg"`
}

type SignUpResponse struct{
	Data Data `json:"data"`	
	Success bool `json:"success"`
}

func GenerateFakeSignUpRequest() *SignUpRequest{
	return &SignUpRequest{
		Username: faker.FirstName(),	
		Password: faker.Password(),
		Mobile: strings.ReplaceAll(faker.Phonenumber(), "-", ""),
	}
}

func BulkSignUp(count, bulkCount int) []*SignUpResponse{
	
	responses := make([]*SignUpResponse, 0, count)

	buffer := make(chan int, bulkCount)

	var wg sync.WaitGroup
	var mutex sync.Mutex
	
	for i:=0; i<count; i++{
		buffer<-i
		wg.Add(1)	
		fmt.Printf("%v. Request", i)	
		go func(){

			signUpRequest := SignUp(GenerateFakeSignUpRequest())	

			mutex.Lock()
			responses = append(responses, signUpRequest)	
			mutex.Unlock()

			wg.Done()
			<-buffer
		}()
	}

	wg.Wait()

	return responses
}

func SignUp(request *SignUpRequest) *SignUpResponse{
	bodyBuffer, _ := json.Marshal(request)
	body := bytes.NewBuffer(bodyBuffer)
	client := http.Client{}
	
	log.Println("Request Body:", string(bodyBuffer))
	res, err := client.Post(fmt.Sprintf("%v/user/signUp", BASE_URL), "application/json", body)
	if err != nil{
		log.Println(err.Error())	
		return nil
	}
	
	defer res.Body.Close()

	if res.StatusCode == 200{

		var signUpResponse SignUpResponse
		response, err := io.ReadAll(res.Body)
		if err != nil{
			log.Println(err.Error())	
			return nil
		}

		if err := json.Unmarshal(response, &signUpResponse); err != nil{
			log.Println(err.Error())
			return nil
		}
		log.Println(signUpResponse)
		return &signUpResponse
	}else{
		response, _:= io.ReadAll(res.Body)
		log.Println(string(response))
		return nil
	}
}