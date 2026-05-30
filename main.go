package main

import (
	config "GopherQueue/Internal/Config"
	httpserver "GopherQueue/Internal/Http"
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	var cfgPath string
	flag.StringVar(&cfgPath, "config", "", "path to config file (yaml or json)")
	flag.Parse()

	if cfgPath == "" {
		if v := os.Getenv("CONFIG_PATH"); v != "" {
			cfgPath = v
		} else {
			cfgPath = "config.yaml"
		}
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	srv := httpserver.New(cfg)

	// Запускаем сервер
	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("http server error: %v", err)
		}
	}()

	// Грейсфул завершение
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
	log.Println("stopped")
}
