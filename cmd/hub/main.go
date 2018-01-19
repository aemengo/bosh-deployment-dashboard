package main

import (
	"database/sql"
	"encoding/json"
	"github.com/aemengo/bosh-deployment-dashboard/config"
	"github.com/aemengo/bosh-deployment-dashboard/info"
	"github.com/jmoiron/sqlx"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"time"
)

type Metrics struct {
	ID                 int       `json:"id" db:"id"`
	InstanceID         string    `json:"instance_id" db:"instance_id"`
	Name               string    `json:"name" db:"name"`
	Address            string    `json:"address" db:"address"`
	AZ                 string    `json:"az" db:"az"`
	Deployment         string    `json:"deployment" db:"deployment"`
	InstanceIndex      int       `json:"instance_index" db:"instance_index"`
	IP                 string    `json:"ip" db:"ip"`
	Label              string    `json:"label" db:"label"`
	CpuUsed            float64   `json:"cpu_used" db:"cpu_used"`
	MemoryUsed         float64   `json:"memory_used" db:"memory_used"`
	PersistentDiskUsed float64   `json:"persistent_disk_used" db:"persistent_disk_used"`
	Load15             float64   `json:"load_15" db:"load_15"`
	Uptime             int       `json:"uptime" db:"uptime"`
	UpdatedAt          time.Time `json:"-" db:"updated_at"`
}

func main() {
	logger := log.New(os.Stdout, "[BDD-H] ", log.LstdFlags)

	if len(os.Args) != 2 {
		logger.Fatalf("[USAGE] %s /path/to/config.yml\n", os.Args[0])
	}

	configPath := os.Args[1]

	cfg, err := config.NewConfig(configPath)
	if err != nil {
		logger.Fatalf("Error %s\n", err)
	}

	db, err := sql.Open("sqlite3", cfg.HubDataDir+"/bdd-hub.db")
	if err != nil {
		logger.Printf("Error opening database at %s: %s", cfg.HubDataDir+"/bdd-hub.db", err)
	}

	dbClient := sqlx.NewDb(db, "sqlite3")

	dbClient.MustExec(`
	create table if not exists metrics (
	  id integer not null primary key,
	  instance_id text unique,
	  name text,
	  address text,
	  az text,
	  deployment text,
	  instance_index integer,
	  ip text,
	  label text,
	  cpu_used real,
	  memory_used real,
	  persistent_disk_used real,
	  load_15 real,
	  uptime integer,
	  updated_at timestamp default current_timestamp not null
	);
	`)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetHealth(w, r, dbClient, logger)
		case http.MethodPost:
			handlePostHealth(w, r, dbClient, logger)
		}
	})

	logger.Printf("Initializing hub on addr: %s\n", cfg.HubAddr)
	logger.Fatal(http.ListenAndServe(cfg.HubAddr, nil))
}

func handleGetHealth(w http.ResponseWriter, r *http.Request, dbClient *sqlx.DB, logger *log.Logger) {
	metrics, err := getMetricsFromDB(dbClient)
	if err != nil {
		logger.Printf("Error retrieving system information from DBs: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func handlePostHealth(w http.ResponseWriter, r *http.Request, dbClient *sqlx.DB, logger *log.Logger) {
	var i info.Info

	if err := json.NewDecoder(r.Body).Decode(&i); err != nil {
		logger.Printf("Error reading json body of request: %s\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := writeInfoToDB(dbClient, i); err != nil {
		logger.Printf("Error writing system information to db for %s: %s\n", i.Spec.ID, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("success"))
}

func getMetricsFromDB(dbClient *sqlx.DB) (metrics []Metrics, err error) {
	err = dbClient.Select(&metrics, "select * from metrics")
	return
}

func writeInfoToDB(dbClient *sqlx.DB, systemInfo info.Info) error {
	_, err := dbClient.Exec(`
	insert or replace into metrics (
	  instance_id,
	  name,
	  address,
	  az,
	  deployment,
	  instance_index,
	  ip,
	  label,
	  cpu_used,
	  memory_used,
	  persistent_disk_used,
	  load_15,
	  uptime
	) VALUES (
	  $1,
	  $2,
	  $3,
	  $4,
	  $5,
	  $6,
	  $7,
	  $8,
	  $9,
	  $10,
	  $11,
	  $12,
	  $13
	  )
	`,
		systemInfo.Spec.ID,
		systemInfo.Spec.InstanceName,
		systemInfo.Spec.Address,
		systemInfo.Spec.AZ,
		systemInfo.Spec.Deployment,
		systemInfo.Spec.Index,
		systemInfo.Spec.IP,
		systemInfo.Label,
		systemInfo.Stats.CpuUsed,
		systemInfo.Stats.MemoryUsed,
		systemInfo.Stats.PersistentDiskUsed,
		systemInfo.Stats.Load15,
		systemInfo.Stats.Uptime,
	)

	return err
}
