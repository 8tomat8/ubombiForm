package environment

import (
	"github.com/kelseyhightower/envconfig"
	"fmt"
	"github.com/dpapathanasiou/go-recaptcha"
	"github.com/jinzhu/gorm"
	"github.com/garyburd/redigo/redis"
	"sync"
	"time"
	h "github.com/8tomat8/ubombiForm/helpers"
	"errors"
	"log"
)

type empty struct{}

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
	DebugDB       bool
	StopTimeout   int `default:"60"`
}

type Env struct {
	DB        *gorm.DB
	Conf      *config
	RedisConn redis.Conn

	wg   *sync.WaitGroup
	stop chan empty
}

func (env *Env) parseConfig() error {
	env.Conf = &config{}
	err := envconfig.Process("ubombiForm", env.Conf)
	if err != nil {
		return err
	}
	format := "MysqlHost: %s\nMysqlPort: %s\nMysqlDB: %s\nMysqlUser: %s\nMysqlPass: %s\nRedisHost: %s\nRedisPort: %s\nRecaptchaKey: %s\nServerHost: %s\nServerPort: %s\n"
	_, err = fmt.Printf(format,
		env.Conf.MysqlHost,
		env.Conf.MysqlPort,
		env.Conf.MysqlDB,
		env.Conf.MysqlUser,
		env.Conf.MysqlPass,
		env.Conf.RedisHost,
		env.Conf.RedisPort,
		env.Conf.RecaptchaKey,
		env.Conf.ServerHost,
		env.Conf.ServerPort)
	if err != nil {
		return err
	}
	return nil
}

func (env *Env) recaptchaInit() {
	recaptcha.Init(env.Conf.RecaptchaKey)
}

func (env *Env) dbInit() (err error) {
	env.DB, err = gorm.Open("mysql",
		fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
			env.Conf.MysqlUser,
			env.Conf.MysqlPass,
			env.Conf.MysqlHost,
			env.Conf.MysqlPort,
			env.Conf.MysqlDB))

	if err != nil {
		return err
	}

	// SQL queries tracing
	env.DB.LogMode(env.Conf.DebugDB)

	env.wg.Add(1)
	go func() {
		<-env.stop
		env.DB.Close()
		env.wg.Done()
	}()
	return nil
}

func (env *Env) redisInit() (err error) {
	env.RedisConn, err = redis.Dial("tcp",
		fmt.Sprintf("%s:%s", env.Conf.RedisHost, env.Conf.RedisPort))
	if err != nil {
		return err
	}
	env.wg.Add(1)
	go func() {
		<-env.stop
		env.RedisConn.Close()
		env.wg.Done()
	}()
	return nil
}

func (env *Env) Start() (err error) {
	env.wg = &sync.WaitGroup{}
	env.stop = make(chan empty)

	err = env.parseConfig()
	if h.Check(err) {
		return err
	}

	err = env.dbInit()
	if h.Check(err) {
		return err
	}

	err = env.redisInit()
	if h.Check(err) {
		return err
	}

	env.recaptchaInit()

	return nil
}

func (env *Env) Stop() error {
	close(env.stop)

	stop := make(chan empty)
	go func() {
		defer close(stop)
		env.wg.Wait()
	}()

	for {
		select {
		case <-stop:
			log.Println("Environment was successfully shut down.")
			return nil
		case <-time.NewTimer(time.Second * time.Duration(env.Conf.StopTimeout)).C:
			return errors.New("Application was stoped by timeout!")
		}
	}

}
