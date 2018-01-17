package main

import (
	"log"
	"os"
	"github.com/aemengo/bosh-deployment-dashboard/config"
	"os/signal"
	"syscall"
	"time"
	"fmt"
	"net/http"
	"encoding/json"
	"bytes"
)

type Metrics struct {
	Spec config.Spec `json:"spec"`
	Label string `json:"label"`
}

func main() {
	logger := log.New(os.Stdout, "[BDD-A] ", log.LstdFlags)

	if len(os.Args) != 2 {
		logger.Fatalf("[USAGE] %s /path/to/config.yml\n", os.Args[0])
	}

	configPath := os.Args[1]

	cfg, err := config.NewConfig(configPath)
	if err != nil {
		logger.Fatalf("Error %s\n", err)
	}

	tickerChan := time.NewTicker(10 * time.Second)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-tickerChan.C:
			sendHealthMetrics(cfg, logger)
		case <-signalChan:
			logger.Println("Shutting down now...")
			return
		}
	}
}

func sendHealthMetrics(cfg config.Config, logger *log.Logger) {
	metrics := Metrics{
		Spec: cfg.Spec,
		Label: cfg.Label,
	}

	contents, _ := json.Marshal(metrics)
	url := fmt.Sprintf("http://%s/health", cfg.HubAddr)
	_ , err := http.Post(url, "application/json", bytes.NewReader(contents))
	if err != nil {
		logger.Printf("Error sending metrics to hub at: %s: %s", cfg.HubAddr, err)
	}
}
