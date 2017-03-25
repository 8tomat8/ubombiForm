package main

import (
	"net/http"
	"log"
	"encoding/json"
	"io/ioutil"
)

type Resp struct {
	Error string  `json:"error"`
	Data  []Count `json:"data"`
}

type Count struct {
	RegionID string `json:"region_id"`
	Count    int    `json:"count"`
}

func check(err error) bool {
	if err != nil {
		log.Println(err)
		return true
	}
	return false
}

func GetStats(w http.ResponseWriter, _ *http.Request) {
	var count []Count
	wJson := json.NewEncoder(w)

	//reply, err := redis.Values(redisConn.Do("HGETALL", "htest"))
	//
	//var results []struct {
	//	Key string
	//	Count int
	//}
	//
	//if err := redis.ScanSlice(reply, &results); err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Printf("%v\n",results)
	//return

	rows, err := db.Table("votes").Select("count(*) as value, region_id").Group("region_id").Rows()
	if check(err) {
		w.WriteHeader(http.StatusInternalServerError)
		err = wJson.Encode(Resp{Error: err.Error()})
		check(err)
		return
	}

	var c Count
	for rows.Next() {
		rows.Scan(&c.Count, &c.RegionID)
		count = append(count, c)
	}

	//err = wJson.Encode(v)
	err = wJson.Encode(Resp{Data: count})
	check(err)
}

func AddVote(w http.ResponseWriter, r *http.Request) {
	wJson := json.NewEncoder(w)

	var request struct {
		Vote
		RecaptchaResp string `json:"g-recaptcha-response"`
	}

	body, err := ioutil.ReadAll(r.Body)
	r.Body.Close()

	log.Println(string(body))
	err = json.Unmarshal(body, &request)
	if check(err) {
		w.WriteHeader(http.StatusBadRequest)
		err := wJson.Encode(Resp{Error: err.Error()})
		check(err)
		return
	}

	log.Printf("%+v\n", request)

	// Check captcha
	if !checkReCaptcha(request.RecaptchaResp, r.RemoteAddr) {
		w.WriteHeader(http.StatusBadRequest)
		err := wJson.Encode(Resp{Error: "Captcha validation failed"})
		check(err)
		return
	}

	// Save to DB
	err = db.Create(&request.Vote).Error
	if check(err) {
		w.WriteHeader(http.StatusInternalServerError)
		err := wJson.Encode(Resp{Error: err.Error()})
		check(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	err = wJson.Encode(Resp{})
	check(err)
}

func GetRegions(w http.ResponseWriter, _ *http.Request) {
	var regions []Region
	wJson := json.NewEncoder(w)

	err := db.Find(&regions).Error
	if check(err) {
		w.WriteHeader(http.StatusInternalServerError)
		err = wJson.Encode(Resp{Error: err.Error()})
		check(err)
		return
	}
	err = wJson.Encode(regions)
	check(err)
}