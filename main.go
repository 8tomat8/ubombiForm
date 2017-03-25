package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"log"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/dpapathanasiou/go-recaptcha"
	"github.com/garyburd/redigo/redis"
	"github.com/kelseyhightower/envconfig"
	"fmt"
)

var db *gorm.DB
var redisConn redis.Conn
var conf *config

type config struct {
	MysqlHost     string
	MysqlPort     string
	MysqlDB       string
	MysqlUser     string
	MysqlPass     string
	RedisHost     string
	RedisPort     string
	RecaptchaKey  string
	ServerHost    string
	ServerPort    string
	IgnoreCaptcha bool
}

func main() {
	var err error

	err = parseConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Init DB
	db, err = gorm.Open("mysql", fmt.Sprintf("%s:@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", conf.MysqlUser, conf.MysqlHost, conf.MysqlPort, conf.MysqlDB))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// SQL queries tracing
	db.LogMode(true)

	// Init Redis
	redisConn, err = redis.Dial("tcp", fmt.Sprintf("%s:%s", conf.RedisHost, conf.RedisPort))
	if err != nil {
		log.Fatal(err)
	}

	// Init reCaptcha
	recaptcha.Init(conf.RecaptchaKey)

	// Init Routes
	r := mux.NewRouter()
	r.Handle("/", http.FileServer(http.Dir("")))
	r.HandleFunc("/regions", GetRegions).Methods("GET")
	r.HandleFunc("/vote", GetStats).Methods("GET")
	r.HandleFunc("/vote", AddVote).Methods("POST")

	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%s", conf.ServerHost, conf.ServerPort), r))
}

func parseConfig() error {
	var c config
	err := envconfig.Process("ubombiForm", &c)
	if err != nil {
		return err
	}
	format := "MysqlHost: %s\nMysqlPort: %s\nMysqlDB: %s\nMysqlUser: %s\nMysqlPass: %s\nRedisHost: %s\nRedisPort: %s\nRecaptchaKey: %s\nServerHost: %s\nServerPort: %s\n"
	_, err = fmt.Printf(format, c.MysqlHost, c.MysqlPort, c.MysqlDB, c.MysqlUser, c.MysqlPass, c.RedisHost, c.RedisPort, c.RecaptchaKey, c.ServerHost, c.ServerPort)
	if err != nil {
		return err
	}
	conf = &c
	return nil
}
