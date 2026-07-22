package main

import (
	"net/http"

	apihandler "github.com/devxdh/pulse/pkg/handler"
)

func (app *application) routes(mux *http.ServeMux) {
	api := apihandler.New(app.db)
	mux.HandleFunc("GET /", api.HomeHandler)
	mux.HandleFunc("POST /api/job", api.CreateJob)
}
