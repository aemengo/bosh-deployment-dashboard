package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aemengo/bosh-deployment-dashboard/config"
	"github.com/aemengo/bosh-deployment-dashboard/system"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Info struct {
	Spec  config.Spec  `json:"spec"`
	Label string       `json:"label"`
	Stats system.Stats `json:"system_stats"`
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
			sendVMInformation(cfg, logger)
		case <-signalChan:
			logger.Println("Shutting down now...")
			return
		}
	}
}

func sendVMInformation(cfg config.Config, logger *log.Logger) {
	stats, err := system.GetStats()
	if err != nil {
		logger.Printf("Error retrieving system level stats: %s", err)
		return
	}

	info := Info{
		Spec:  cfg.Spec,
		Label: cfg.Label,
		Stats: stats,
	}

	contents, _ := json.Marshal(info)
	url := fmt.Sprintf("http://%s/health", cfg.HubAddr)
	response, err := http.Post(url, "application/json", bytes.NewReader(contents))
	if err != nil {
		logger.Printf("Error sending metrics to hub at: %s: %s", cfg.HubAddr, err)
		return
	}

	if response.StatusCode != http.StatusOK {
		logger.Printf("Failed sending metrics to hub at: %s: %s", cfg.HubAddr, response.Status)
		return
	}
}
