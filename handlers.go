package main

import (
	"net/http"
	"encoding/json"
	"io/ioutil"
	"github.com/8tomat8/ubombiForm/environment"
	"net"
	"github.com/dpapathanasiou/go-recaptcha"
	h "github.com/8tomat8/ubombiForm/helpers"
	"github.com/garyburd/redigo/redis"
	"github.com/pkg/errors"
	"github.com/ulule/paging"
)

const cacheKey string = "ubombiForm:cachedVoteStats"

type Resp struct {
	Error string  `json:"error"`
	Data  interface{} `json:"data"`
}

type Count struct {
	RegionID int `json:"region_id"`
	Count    int    `json:"count"`
}

type Handle struct {
	environment.Env
}

func (ha *Handle) GetStats(w http.ResponseWriter, _ *http.Request) {
	wJson := json.NewEncoder(w)

	// Trying to get cached stats from Redis
	data, err := ha.getCachedStats()
	if !h.Check(err) {
		err = wJson.Encode(Resp{Data: data})
		h.Check(err)
		return
	}

	// Trying to get stats from DB
	data, err = ha.getPersistStats()
	if h.Check(err) {
		w.WriteHeader(http.StatusBadRequest)
		err := wJson.Encode(Resp{Error: err.Error()})
		h.Check(err)
		return
	}

	// Caching stats to Redis
	err = ha.setCachedStats(data)
	h.Check(err)

	err = wJson.Encode(Resp{Data: data})
	h.Check(err)
}

func (ha *Handle) GetVotes(w http.ResponseWriter, r *http.Request) {
	var votes []*Vote
	wJson := json.NewEncoder(w)

	store, err := paging.NewGORMStore(ha.DB, &votes)
	if h.Check(err) {
		w.WriteHeader(http.StatusInternalServerError)
		err = wJson.Encode(Resp{Error: err.Error()})
		h.Check(err)
		return
	}

	// TODO: singleton
	opts := paging.NewOptions()
	opts.OffsetKeyName = "page"
	opts.LimitKeyName = "on_page"

	paginator, err := paging.NewOffsetPaginator(store, r, opts)
	if h.Check(err) {
		w.WriteHeader(http.StatusInternalServerError)
		err = wJson.Encode(Resp{Error: err.Error()})
		h.Check(err)
		return
	}

	err = paginator.Page()
	if h.Check(err) {
		w.WriteHeader(http.StatusInternalServerError)
		err = wJson.Encode(Resp{Error: err.Error()})
		h.Check(err)
		return
	}

	err = wJson.Encode(votes)
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
	if !ha.checkReCaptcha(request.RecaptchaResp, r.RemoteAddr) {
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

func (ha *Handle) checkReCaptcha(recaptchaResponse string, remoteAddr string) (result bool) {
	if ha.Conf.IgnoreCaptcha {
		return true
	}
	ip, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return
	}
	result = recaptcha.Confirm(ip, recaptchaResponse)
	return
}

func (ha *Handle) getCachedStats() (counts []interface{}, err error) {
	conn := ha.RedisPool.Get()
	defer conn.Close()
	reply, err := redis.Values(conn.Do("HGETALL", cacheKey))
	if h.Check(err) {
		return
	}

	for len(reply) > 0 {
		c := &Count{}
		reply, err = redis.Scan(reply, &c.RegionID, &c.Count)
		if h.Check(err) {
			return
		}
		counts = append(counts, c)
	}
	if len(counts) == 0 {
		return nil, errors.New("No cached stats in Redis.")
	}

	return
}

func (ha *Handle) getPersistStats() (counts []interface{}, err error) {
	rows, err := ha.DB.Table("votes").Select("count(*) as count, region_id").Group("region_id").Rows()
	if h.Check(err) {
		return
	}

	for rows.Next() {
		c := &Count{}
		err = rows.Scan(&c.Count, &c.RegionID)
		if h.Check(err) {
			return
		}
		counts = append(counts, c)
	}
	return
}

func (ha *Handle) setCachedStats(counts []interface{}) (err error) {
	conn := ha.RedisPool.Get()
	defer conn.Close()

	for _, count := range counts {
		err = conn.Send("HSET", cacheKey, count.(*Count).RegionID, count.(*Count).Count)
		if h.Check(err) {
			return err
		}
	}

	err = conn.Send("EXPIRE", cacheKey, ha.Conf.RedisCacheTTL)
	if h.Check(err) {
		return err
	}

	return nil
}
