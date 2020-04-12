package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

// we store a map of buckets for each user by api key
var visitors = make(map[string]*Bucket)

// a bucket represents available connex, basically
type Bucket struct {
	Cap      int
	Interval time.Duration
	PerDrip  int
	consumed int
	started  bool
	kill     chan bool
	m        sync.Mutex
}

func (b *Bucket) StartBucket() error {
	if b.started {
		return errors.New("Bucket already exists")
	}

	// go's ticker sends a message at set intervals
	ticker := time.NewTicker(b.Interval)
	b.started = true
	b.kill = make(chan bool, 1)

	// on every "drip" we reset the requests consumed to 0
	go func() {
		for {
			select {
			case <-ticker.C:
				b.m.Lock()
				b.consumed -= b.PerDrip
				if b.consumed < 0 {
					b.consumed = 0
				}
				b.m.Unlock()
			case <-b.kill:
				return
			}
		}
	}()

	return nil
}

func (b *Bucket) StopBucket() error {
	if !b.started {
		return errors.New("Bucket doesn't exist")
	}

	b.kill <- true

	return nil
}

func (b *Bucket) Consume() error {
	b.m.Lock()
	defer b.m.Unlock()

	// if the bucket is empty return an error
	if b.Cap-b.consumed < 1 {
		return errors.New("Not enough capacity.")
	}
	b.consumed += 1
	return nil
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/limit", func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-API-KEY")
		// if the bucket exists, we've seen the user before
		// otherwise, we haven't and need to make a new bucket
		if b, ok := visitors[key]; ok {
			fmt.Println("using existing bucket for user")
			err := b.Consume()
			if err != nil {
				respondWithError(w, 429, "rate limit exceeded")
			}
		} else {
			fmt.Println("creating bucket for new user")
			b = &Bucket{
				Cap:      10,
				Interval: 1 * time.Second,
				PerDrip:  10,
			}

			b.Consume()
			visitors[key] = b
		}

		respondWithJSON(w, 200, "hit endpoint successfully okokok")
	})

	r.Use(AuthMiddleware)

	fmt.Println("listening on port 3000")
	http.ListenAndServe(":3000", r)
}
