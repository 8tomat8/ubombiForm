package main

import (
	"net/http"
	"encoding/json"
	"io/ioutil"
	"github.com/8tomat8/ubombiForm/environment"
	"net"
	"github.com/dpapathanasiou/go-recaptcha"
	h "github.com/8tomat8/ubombiForm/helpers"
)

type Resp struct {
	Error string  `json:"error"`
	Data  []Count `json:"data"`
}

type Count struct {
	RegionID string `json:"region_id"`
	Count    int    `json:"count"`
}

type Handle struct {
	environment.Env
}

func (ha *Handle) GetStats(w http.ResponseWriter, _ *http.Request) {
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

	rows, err := ha.DB.Table("votes").Select("count(*) as value, region_id").Group("region_id").Rows()
	if h.Check(err) {
		w.WriteHeader(http.StatusInternalServerError)
		err = wJson.Encode(Resp{Error: err.Error()})
		h.Check(err)
		return
	}

	var c Count
	for rows.Next() {
		rows.Scan(&c.Count, &c.RegionID)
		count = append(count, c)
	}

	//err = wJson.Encode(v)
	err = wJson.Encode(Resp{Data: count})
	h.Check(err)
}

func (ha *Handle) AddVote(w http.ResponseWriter, r *http.Request) {
	wJson := json.NewEncoder(w)

	var request struct {
		Vote
		RecaptchaResp string `json:"g-recaptcha-response"`
	}

	body, err := ioutil.ReadAll(r.Body)
	r.Body.Close()

	err = json.Unmarshal(body, &request)
	if h.Check(err) {
		w.WriteHeader(http.StatusBadRequest)
		err := wJson.Encode(Resp{Error: err.Error()})
		h.Check(err)
		return
	}

	// Check captcha
	if !checkReCaptcha(ha, request.RecaptchaResp, r.RemoteAddr) {
		w.WriteHeader(http.StatusBadRequest)
		err := wJson.Encode(Resp{Error: "Captcha validation failed"})
		h.Check(err)
		return
	}

	// Save to DB
	err = ha.DB.Create(&request.Vote).Error
	if h.Check(err) {
		w.WriteHeader(http.StatusInternalServerError)
		err := wJson.Encode(Resp{Error: err.Error()})
		h.Check(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	err = wJson.Encode(Resp{})
	h.Check(err)
}

func (ha *Handle) GetRegions(w http.ResponseWriter, _ *http.Request) {
	var regions []Region
	wJson := json.NewEncoder(w)

	err := ha.DB.Find(&regions).Error
	if h.Check(err) {
		w.WriteHeader(http.StatusInternalServerError)
		err = wJson.Encode(Resp{Error: err.Error()})
		h.Check(err)
		return
	}
	err = wJson.Encode(regions)
	h.Check(err)
}

func checkReCaptcha(h *Handle, recaptchaResponse string, remoteAddr string) (result bool) {
	if h.Conf.IgnoreCaptcha {
		return true
	}
	ip, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return
	}
	result = recaptcha.Confirm(ip, recaptchaResponse)
	return
}
