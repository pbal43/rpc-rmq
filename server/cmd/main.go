package main

import (
	"context"
	"os"
	"os/signal"
	"server/internal"
	"server/internal/server"
	"sync"
	"syscall"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
		defer signal.Stop(c)
		<-c
		cancel()
	}()

	cfg := internal.ReadConfig()

	wg := sync.WaitGroup{}

	srv := server.NewServer(cfg)

	srv.Run()

	wg.Add(1)
	go func() {
		defer wg.Done()
		srv.StartConsume()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		srv.Stop()
	}()

	wg.Wait()
}
