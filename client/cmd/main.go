package main

import (
	"client/internal"
	"client/internal/server"
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	timeOut := time.Second * 5

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
		defer signal.Stop(c)
		<-c
		cancel()
	}()

	cfg := internal.ReadConfig()

	wg := sync.WaitGroup{}

	client := server.NewClient(cfg)

	client.Run()

	wg.Add(1)
	go func() {
		defer wg.Done()
		number := 4
		doubled, err := client.Call(float32(number), timeOut)
		if err != nil {
			log.Printf("client.Call(%d): %v", number, err)
		}

		log.Printf("client.Call(%d): %v", number, doubled)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		number := 11
		doubled, err := client.Call(float32(number), timeOut)
		if err != nil {
			log.Printf("client.Call(%d): %v", number, err)
		}

		log.Printf("client.Call(%d): %v", number, doubled)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		client.Stop()
	}()

	wg.Wait()
}
