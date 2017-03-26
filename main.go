package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"log"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"fmt"
	"github.com/8tomat8/ubombiForm/environment"
	h "github.com/8tomat8/ubombiForm/helpers"
	"context"
	"os/signal"
	"syscall"
	"os"
)

type empty struct{}

func main() {
	var err error
	var env environment.Env

	err = env.Start()
	if err != nil {
		log.Fatal(err)
	}

	handle := Handle{env}

	// Init Routes
	r := mux.NewRouter()
	r.Handle("/", http.FileServer(http.Dir("")))
	r.HandleFunc("/regions", handle.GetRegions).Methods("GET")
	r.HandleFunc("/vote", handle.GetStats).Methods("GET")
	r.HandleFunc("/vote", handle.AddVote).Methods("POST")

	// Init HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", env.Conf.ServerHost, env.Conf.ServerPort),
		Handler: r,
	}

	// Graceful shutdown
	stop := make(chan empty)
	go func() {
		sigs := make(chan os.Signal)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		log.Printf("Received %v signal\n", <-sigs)
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Fatalf("could not shutdown: %v", err)
		}
		close(stop)
	}()

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
	}
	<-stop
	env.Stop()
	h.Check(err)
}
