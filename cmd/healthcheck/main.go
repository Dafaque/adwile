package main

import (
	"context"
	"flag"
	"healthcheck/internal/checker"
	"healthcheck/internal/config"
	"healthcheck/internal/storage"
	"log"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

func main() {
	configPath := flag.String("c", "config/config.json", "config file path")
	flag.Parse()

	cfg, errOpenConfig := config.NewConfig(*configPath)
	if errOpenConfig != nil {
		log.Fatalf("errOpenConfig: %s", errOpenConfig)
	}
	checker := checker.NewChecker(
		cfg.HttpClientTimeoutSec,
		storage.NewStorageSQLite3(
			cfg.ConnStr,
			cfg.DbOpTimeoutSec,
			cfg.DbEnableStdout,
		))

	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(cfg.CheckTimeoutSec*int(time.Second)))
	defer cancel()

	var doneUrls int64
	var lenUrls = len(cfg.Urls)
	var done chan struct{} = make(chan struct{})
	for _, u := range cfg.Urls {
		go func(urlConfig config.ConfigUrl) {
			checker.Check(ctx, urlConfig)
			if atomic.AddInt64(&doneUrls, 1) == int64(lenUrls) {
				done <- struct{}{}
			}

		}(u)
	}

	var sig chan os.Signal = make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGABRT, syscall.SIGKILL)

	select {
	case <-ctx.Done():
		log.Println("Deadline exceeded")
		return
	case <-sig:
		log.Println("Exited")
		return
	case <-done:
		log.Println("Done")
		return
	}
}
