package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	dbConnect := os.Getenv("SQL_DB")
	if dbConnect == "" {
		log.Fatal("You need to specify a string for connecting to the SQL db")
	}
	err := InitDB(dbConnect)
	if err != nil {
		log.Fatalf("Unable to initalize database: %v", err)
	}

	r := mux.NewRouter()
	r.Methods("OPTIONS").HandlerFunc(corsHandler)

	v1Router := r.PathPrefix("/1").Subrouter()
	installEndpoints(v1Router)

	server := &http.Server{
		Addr:         ":5555",
		Handler:      r,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	log.Printf("Starting insecure (HTTP) server on port 5555")
	log.Fatal(server.ListenAndServe())
}

func installEndpoints(router *mux.Router) {
	router.Handle("/sessions", NewtonFunc(CreateSessionHandler)).Methods("POST")

	router.Handle("/users", NewtonFunc(CreateUserHandler)).Methods("POST")
	router.Handle("/users/{user_id}", NewtonFunc(GetUserHandler)).Methods("GET")
	router.Handle("/users/{user_id}", NewtonFunc(EditUserHandler)).Methods("PUT")

	router.Handle("/bookmarks", NewtonFunc(CreateBookmarkHandler)).Methods("POST")
	router.Handle("/bookmarks", NewtonFunc(GetBookmarksHandler)).Methods("GET")
	router.Handle("/bookmarks/{bookmark_id}", NewtonFunc(GetBookmarkHandler)).Methods("GET")
	router.Handle("/bookmarks/{bookmark_id}", NewtonFunc(EditBookmarkHandler)).Methods("PUT")
	router.Handle("/bookmarks/{bookmark_id}", NewtonFunc(DeleteBookmarkHandler)).Methods("DELETE")

	router.Handle("/contacts", NewtonFunc(CreateContactHandler)).Methods("POST")
	router.Handle("/contacts", NewtonFunc(GetContactsHandler)).Methods("GET")
	router.Handle("/contacts/{contact_id}", NewtonFunc(GetContactHandler)).Methods("GET")
	router.Handle("/contacts/{contact_id}", NewtonFunc(DeleteContactHandler)).Methods("DELETE")
}

func corsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", r.Header.Get("Access-Control-Request-Headers"))
	w.WriteHeader(http.StatusOK)
}
