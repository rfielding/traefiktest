package main

/*

 This simple service:

 - responds to pings
 - goes down with a lower probability than it goes up with
 - helps to test load balancing

 */

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

var bind = ""
var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
var stateChan = make(chan bool,0)

func downState(w http.ResponseWriter, r *http.Request) bool {
	if false == <-stateChan {
		w.WriteHeader(500)
		msg := fmt.Sprintf("%s is down", bind)
		w.Write([]byte(msg))
		return true
	}
	return false
}

func getRoot(w http.ResponseWriter, r *http.Request) {
	if downState(w,r) {
		return
	}
 
	w.Write([]byte(fmt.Sprintf("server %s\n", bind)))
}

func getPing(w http.ResponseWriter, r *http.Request) {
	if downState(w,r) {
		return
	}

	w.Write([]byte("pong"))
	fmt.Sprintf("pong")
}

func main() {
	flag.StringVar(&bind, "bind", "localhost:1111", "the bind address for this service")
	flag.Parse()
	http.HandleFunc("/", getRoot)
	http.HandleFunc("/ping", getPing)

	// Randomly oscillate between up and down state
	go func() {
		// Sit in a loop randomly changing state from up to down
		isUp := false
		for {
			if isUp && rnd.Intn(10) < 1 {
				isUp = !isUp
			} else if !isUp && rnd.Intn(10) < 7 {
				isUp = !isUp
			}
			
			stateChan <- isUp
		}
	}() 

	fmt.Printf("Serve on: %s\n", bind)
	err := http.ListenAndServe(bind, nil)
	if err != nil {
		panic(fmt.Sprintf("http server died: %v\n", err))
	}
}
