package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aemengo/bosh-deployment-dashboard/cf"
	"github.com/aemengo/bosh-deployment-dashboard/config"
	"github.com/aemengo/bosh-deployment-dashboard/info"
	"github.com/aemengo/bosh-deployment-dashboard/system"
)

func main() {
	logger := log.New(os.Stdout, "[BDD-A] ", log.LstdFlags)

	if len(os.Args) != 2 {
		logger.Fatalf("[USAGE] %s /path/to/config.yml", os.Args[0])
	}

	configPath := os.Args[1]

	cfg, err := config.NewConfig(configPath)
	if err != nil {
		logger.Fatalf("Error: %s", err)
	}

	cf := cf.New(cfg)

	tickerChan := time.NewTicker(10 * time.Second)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-tickerChan.C:
			sendVMInformation(cf, cfg, logger)
		case <-signalChan:
			logger.Println("Shutting down now...")
			return
		}
	}
}

func sendVMInformation(cf *cf.Cf, cfg config.Config, logger *log.Logger) {
	logger.Println("Retrieving system level stats...")
	stats, err := system.GetStats()
	if err != nil {
		logger.Printf("Error retrieving system level stats: %s", err)
		return
	}

	logger.Println("Retrieving cloudfoundry service instance deployment info...")
	deploymentInfo, err := cf.GetDeploymentInfo()
	if err != nil {
		logger.Printf("Error retrieving cloud foundry : %s", err)
		return
	}

	//TODO make^ above run independently at different intervals
	//TODO Also make sure that cf is enabled first

	i := info.Info{
		Spec:  cfg.Spec,
		Label: cfg.Label,
		Stats: stats,
		Cf:    deploymentInfo,
	}

	contents, _ := json.Marshal(i)

	url := fmt.Sprintf("http://%s/api/health", cfg.Hub.Addr())
	response, err := http.Post(url, "application/json", bytes.NewReader(contents))
	if err != nil {
		logger.Printf("Error sending metrics to hub at: %s: %s", cfg.Hub.Addr(), err)
		return
	}

	if response.StatusCode != http.StatusOK {
		logger.Printf("Failed sending metrics to hub at: %s: %s", cfg.Hub.Addr(), response.Status)
		return
	}
}
