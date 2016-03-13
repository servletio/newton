package main

import "net/http"

// LocationRecord ...
type LocationRecord struct {
	Timestamp int64   `json:"timestamp"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	OwnerID   int64   `json:"owner_id"db:"owner_id"`
}

// CreateLocationEntry handles POST /locations
func CreateLocationEntry(w http.ResponseWriter, r *http.Request) {

}
