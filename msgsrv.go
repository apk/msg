package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"fmt"
	"time"
	"strings"
	"encoding/json"
	"os"
)

type msg struct {
	ts time.Time
	data string
	addr []string
}

func compute(ch chan msg) {

	for {
		s := <-ch
		t := s.ts
		yy, mn, dy := t.Date()
		hh, mm, ss := t.Clock()
		fn := fmt.Sprintf("log-%04d-%02d-%02d.txt", yy, mn, dy)
		ln := fmt.Sprintf("%02d:%02d:%02d: %s\n", hh, mm, ss, s.data)
		fmt.Print(ln)
		f, err := os.OpenFile(fn, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
		if err == nil {
			f.Write([]byte(ln))
			f.Close()
		} else {
			fmt.Printf("err=%v\n",err)
		}

		b, err := json.Marshal(s)
		if err == nil {
			fmt.Printf("b=%v\n", string(b));
		}
	}
}

func main() {

	ch := make(chan msg)

	go compute(ch)

	http.Handle("/", http.FileServer(http.Dir(".")))

	http.HandleFunc("/msg/in", func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err == nil {
			ch <- msg{ts: time.Now(), data: string(body), addr: []string{}}
		}

	})

	http.Handle("/msg/in/", http.StripPrefix("/msg/in/",http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err == nil {
			a := strings.Split(r.URL.Path, "/")
			ch <- msg{ts: time.Now(), data: string(body), addr: a}
		}

	})))

	log.Fatal(http.ListenAndServe(":3047", nil))
}
