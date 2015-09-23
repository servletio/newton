package main

import (
	crand "crypto/rand"
	"log"
	"math/big"
	"math/rand"
	"net/http"
	"runtime"
	"strings"
)

type newtonFunc struct {
	f http.HandlerFunc
}

func (nf *newtonFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// enable HSTS
	w.Header().Set("Strict-Transport-Security", "max-age=15768000")

	// catch any panics that occur during execution
	defer func() {
		if r := recover(); r != nil {
			s := make([]byte, 2048)
			numBytes := runtime.Stack(s, false)
			stack := s[:numBytes]
			log.Printf("recovered - %v\n%s", r, string(stack))
		}
	}()

	nf.f(w, r)
}

// NewtonFunc wraps an http.HandlerFunc into our custom http.Handler
func NewtonFunc(f func(http.ResponseWriter, *http.Request)) http.Handler {
	return &newtonFunc{f: f}
}

var gAlphaNums = strings.Split("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", "")

func randAlphaNum(length int) string {
	s := ""
	numRunes := big.NewInt(int64(len(gAlphaNums)))
	for i := 0; i < length; i++ {
		idx, err := crand.Int(crand.Reader, numRunes)
		if err != nil {
			// fall back to non-crypto rand
			s += gAlphaNums[rand.Intn(len(gAlphaNums))]
			continue
		}
		s += gAlphaNums[idx.Int64()]
	}

	return s
}

func logErr(err error) {
	log.Print(err)
}
