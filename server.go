package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

// App represents the server's internal state.
// It holds configuration about providers and content
type App struct {
	Service Service
}

func (a App) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Printf("%s %s", req.Method, req.URL.String())
	limit, offset, err := getParameters(w, req)
	if err != nil {
		writeValidationErrorResponse(w, err)
	}
	items, err := a.Service.ContentItems(limit, offset)
	if err != nil {
		if _, ok := err.(ValidationError); ok {
			writeValidationErrorResponse(w, err)
		} else {
			writeInternalServerErrorResponse(w, err)
		}
	}
	bb, err := json.Marshal(items)
	if err != nil {
		writeInternalServerErrorResponse(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(bb); err != nil {
		log.Println("error when trying to write data to HTTP response: " + err.Error())
	}
	return
}

func getParameters(w http.ResponseWriter, req *http.Request) (limit, offset int, err error) {
	limits, _ := req.URL.Query()["count"]
	if len(limits) != 0 {
		limit, err = strconv.Atoi(limits[0])
		if err != nil {
			return
		}
	}

	offsets, _ := req.URL.Query()["offset"]
	if len(offsets) != 0 {
		offset, err = strconv.Atoi(offsets[0])
		if err != nil {
			return
		}
	}
	return
}

func writeInternalServerErrorResponse(w http.ResponseWriter, err error) {
	log.Print("internal server error: " + err.Error())
	w.WriteHeader(http.StatusInternalServerError)
	if _, err := w.Write([]byte("internal server error: " + err.Error())); err != nil {
		log.Println("error when trying to write data to HTTP response: " + err.Error())
	}
}

func writeValidationErrorResponse(w http.ResponseWriter, err error) {
	log.Print("validation error: " + err.Error())
	w.WriteHeader(http.StatusBadRequest)
	if _, err := w.Write([]byte("invalid input parameters: " + err.Error())); err != nil {
		log.Println("error when trying to write data to HTTP response: " + err.Error())
	}
}
