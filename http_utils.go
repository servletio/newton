package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"runtime"
	"strconv"
)

// NewtonError is used to identify errors that occur when the API returns something
// other than an http 200
type NewtonError int

// Newton API errors
const (
	ErrorNone NewtonError = iota
	ErrorInternal
	ErrorNotFound
	ErrorBadRequest
	ErrorUnauthorized
)

func sendResponse(w http.ResponseWriter, response interface{}, httpCode int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(httpCode)

	enc := json.NewEncoder(w)
	err := enc.Encode(response)
	if err != nil {
		panic(err)
	}
}

func sendErr(w http.ResponseWriter, msg string, httpCode int, apiCode NewtonError) {
	sendResponse(
		w,
		map[string]interface{}{
			"error_message": msg,
			"error_code":    apiCode,
		},
		httpCode)
}

func sendSuccess(w http.ResponseWriter, response interface{}) {
	if response == nil {
		response = map[string]interface{}{}
	}

	sendResponse(w, response, http.StatusOK)
}

func sendNotFound(w http.ResponseWriter, msg string) {
	sendErr(w, msg, http.StatusNotFound, ErrorNotFound)
}

func sendInternalErr(w http.ResponseWriter, err error) {
	sendErr(w, "Internal server error", http.StatusInternalServerError, ErrorInternal)

	if err != nil {
		_, file, line, ok := runtime.Caller(1)
		if !ok {
			file = "???"
			line = 0
		}
		file = filepath.Base(file)
		log.Printf("%s:%d %v", file, line, err)
	}
}

func sendBadReqCode(w http.ResponseWriter, msg string, apiCode NewtonError) {
	sendErr(w, msg, http.StatusBadRequest, apiCode)
}

func sendBadReq(w http.ResponseWriter, msg string) {
	sendBadReqCode(w, msg, ErrorBadRequest)
}

func sendUnauthorized(w http.ResponseWriter, msg string) {
	sendErr(w, msg, http.StatusUnauthorized, ErrorUnauthorized)
}

func pageAndSize(args url.Values, defaultPageSize int) (page, pageSize int, err error) {
	page = 0
	pageSize = defaultPageSize
	if args.Get("page_size") != "" {
		pageSize, err = strconv.Atoi(args.Get("page_size"))
		if err != nil {
			err = errors.New("unable to parse 'page_size'")
			return
		}
		if pageSize < 1 {
			err = fmt.Errorf("page_size must be at least 1, found %d", pageSize)
		}
	}
	if args.Get("page") != "" {
		page, err = strconv.Atoi(args.Get("page"))
		if err != nil {
			err = errors.New("unable to parse 'page'")
			return
		}
		if page < 0 {
			err = fmt.Errorf("page must be at least 0, found %d", page)
			return
		}
	}

	return
}
