package main

import (
	"net/http"
)

func (app *application) routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /", app.api.HomeHandler)
	mux.HandleFunc("POST /api/job", app.api.CreateJob)
	mux.HandleFunc("GET /api/job/{id}", app.api.GetJobByID)
}
