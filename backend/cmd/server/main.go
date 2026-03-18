package main

import (
	"log"
	"net/http"

	"home-decision/backend/internal/config"
	"home-decision/backend/internal/httpapi"
	"home-decision/backend/internal/service"
	"home-decision/backend/internal/store"
)

func main() {
	cfg := config.Load()

	meta := service.DefaultMeta()
	var dataStore store.Store
	var closeFn func() error

	switch cfg.StoreMode {
	case "mysql":
		mysqlStore, err := store.NewMySQLStore(cfg.MySQLDSN, meta)
		if err != nil {
			log.Fatal(err)
		}
		dataStore = mysqlStore
		closeFn = mysqlStore.Close
	default:
		profiles := service.DefaultProfiles()
		houses := service.DefaultHouses()
		dataStore = store.NewMemoryStore(meta, profiles, houses)
	}

	if closeFn != nil {
		defer closeFn()
	}

	scoring := service.NewScoringService(dataStore)
	auth := service.NewAuthService(dataStore)
	server := httpapi.NewServer(dataStore, scoring, auth, cfg.AllowedOrigin)

	log.Printf("home decision backend listening on :%s mode=%s", cfg.Port, cfg.StoreMode)
	if err := http.ListenAndServe(":"+cfg.Port, server.Routes()); err != nil {
		log.Fatal(err)
	}
}
