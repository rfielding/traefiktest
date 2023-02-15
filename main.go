package main

import (
  "net/http"
  "flag"
  "fmt"
)
 
var bind = ""

func getRoot(w http.ResponseWriter, r *http.Request) {
  w.Write([]byte(fmt.Sprintf("server %s\n", bind)))  
}

func getPing(w http.ResponseWriter, r *http.Request) {
  w.Write([]byte("pong"))
}

func main() {
  flag.StringVar(&bind, "bind","localhost:1111","the bind address for this service")
  flag.Parse()
  http.HandleFunc("/",getRoot)
  http.HandleFunc("/ping",getPing)

  fmt.Printf("Serve on: %s", bind)
  err := http.ListenAndServe(bind,nil)
  if err != nil {
    panic(fmt.Sprintf("http server died: %v\n", err))
  }
}
