package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	_ "net/url"
	"strings"
	"time"

	goquery "github.com/PuerkitoBio/goquery"
	request "github.com/mozillazg/request"
	cron "github.com/robfig/cron"
)

func main() {
	log.Println("begin order task ...")
	c := cron.New()
	spec := "1 1 10 * * ?"
	c.AddFunc(spec, saveOrder)
	c.Start()
	select {}
}

func getRandom(randomList []string) string {
	size := len(randomList)
	rand.Seed(time.Now().UnixNano())
	index := rand.Intn(size)
	return randomList[index]
}

type Result struct {
	Code   string
	Data   Data
	Result string
	Status int
}

type Data struct {
	Members string
	Address string
}

type Food struct {
	Code   string
	Data   string
	Result string
	Status int
}

func saveOrder() {

	//login
	req := request.NewRequest(new(http.Client))
	req.Data = map[string]string{
		"LoginForm[username]":  "18357118527",
		"LoginForm[password]":  "74dc7108dc671dc5b3b38c493cbcc4df",
		"LoginForm[autoLogin]": "1",
		"yt0": "登录",
	}
	loginURL := "http://wos.chijidun.com/login.html"

	result, _ := req.Post(loginURL)
	resp := result.Response
	defer resp.Body.Close()

	//get order info
	nowStr := time.Now().Format("2006-01-02")
	orderURL := "http://wos.chijidun.com/order/getMembersAndOrder.html?cid=1648&date=" + nowStr + "&mealType=3"
	orderResp, _ := req.Get(orderURL)
	defer orderResp.Body.Close()
	oBody, _ := ioutil.ReadAll(orderResp.Body)

	var r Result
	err := json.Unmarshal(oBody, &r)
	if err != nil {
		log.Println(err)
	}
	log.Printf("result: %s, %s ,%d \n", r.Result, r.Code, r.Status)

	if r.Code != "200" || r.Status != 1 {
		log.Println("Get Order Info Error:", r.Result)
		return
	}

	data := r.Data
	log.Println("data:", data)

	//parse order info
	var idList []string
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(data.Members))
	doc.Find("ul").Find("li").Each(func(i int, s *goquery.Selection) {
		log.Println("i: ", i)
		var id string
		id, ok := s.Attr("data-id")
		if !ok {
			log.Println("Error When Get CanTing id")
			return
		}
		//log.Println("li content:", id)
		idList = append(idList, id)
	})

	log.Println("id list:", idList)
	randomId := getRandom(idList)
	log.Println("random id:", randomId)

	//get food
	foodURL := ("http://wos.chijidun.com/order/getMenu.html?mid=" + randomId + "&date=" + nowStr + "&type=3")

	foodResp, _ := req.Get(foodURL)
	defer foodResp.Body.Close()
	foodBody, _ := ioutil.ReadAll(foodResp.Body)
	log.Println("body :", string(foodBody))

	var food Food
	foodErr := json.Unmarshal(foodBody, &food)
	if foodErr != nil {
		log.Println("Error :", foodErr)
	}

	if food.Status != 1 && food.Code != "200" {
		log.Println("Error Get Food Info")
	}

	var foodIdList []string
	foodDoc, _ := goquery.NewDocumentFromReader(strings.NewReader(food.Data))
	foodDoc.Find("li").Each(func(i int, s *goquery.Selection) {
		foodId, _ := s.Attr("data-id")
		log.Println("Food Name:", foodId)
		foodIdList = append(foodIdList, foodId)
	})

	foodRandomId := getRandom(foodIdList)
	log.Println("food random Id:", foodRandomId)

	//create a food order
	saveOrderURL := ("http://wos.chijidun.com/order/saveOrder.html")
	req.Data = map[string]string{
		"items":    foodRandomId + ":1;",
		"addrId":   "30",
		"mealType": "3",
		"date":     nowStr,
	}

	_, saveError := req.Post(saveOrderURL)
	if saveError != nil {
		log.Println("Save Order Error: ", saveError)
	}

}
