package handler

import (
	"net/http"

	"taskflow/internal/api"
	"taskflow/internal/store"
)

var server *api.Server

func init() {
	dataFile := store.ResolveFile("TASKFLOW_DATA_FILE", ".")
	dataStore, err := store.New(dataFile)
	if err != nil {
		panic(err)
	}
	server = api.New(dataStore, "public", nil)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	server.Handler().ServeHTTP(w, r)
}
