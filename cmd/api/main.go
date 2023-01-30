package main

import (
	"encoding/json"
	"flag"
	"healthcheck/internal/config"
	"healthcheck/internal/storage"
	"log"
	"net/http"
	"strconv"
)

// я уже решил сделать все масимально просто, в один файл, без доп конфигов и прочих сладостей.

func main() {
	api := NewApi()
	log.Println(api.ListenAndServe())
}

func NewApi() Api {
	var api Api
	{
		configPath := flag.String("c", "config/config.json", "config file path")
		flag.Parse()
		cfg, errOpenConfig := config.NewConfig(*configPath)
		if errOpenConfig != nil {
			log.Fatalf("errOpenConfig: %s", errOpenConfig)
		}
		api.cfg = cfg
		api.saver = storage.NewStorageSQLite3(cfg.ConnStr, cfg.DbOpTimeoutSec, false)
		if api.saver == nil {
			panic("DB is not opened")
		}
	}
	api.Server = http.Server{
		Addr: ":8080",
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/statuses", api.getStatuses)
	api.Handler = mux

	return api
}

type Api struct {
	http.Server
	cfg   *config.Config
	saver *storage.StorageSQLite3
}

func (a *Api) getStatuses(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var offset int
	sOffset := r.URL.Query().Get("offset")
	if len(sOffset) > 0 {
		offset, _ = strconv.Atoi(sOffset)
	}

	result, err := a.saver.GetTopStatuses(offset)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error())) //Непродовая затычка
		return
	}

	b, err := json.Marshal(result)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error())) //Непродовая затычка
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(b)
}
