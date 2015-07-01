package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"fmt"
	"time"
	"os"
)

func compute(ch chan string) {

	for {
		s := <-ch
		t := time.Now()
		yy, mn, dy := t.Date()
		hh, mm, ss := t.Clock()
		fn := fmt.Sprintf("log-%04d-%02d-%02d.txt", yy, mn, dy)
		ln := fmt.Sprintf("%02d:%02d:%02d: %s\n", hh, mm, ss, s)
		fmt.Print(ln)
		f, err := os.OpenFile(fn, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
		if err == nil {
			f.Write([]byte(ln))
			f.Close()
		} else {
			fmt.Println("err=%v\n",err)
		}
	}
}

func main() {

	ch := make(chan string)

	go compute(ch)

	http.Handle("/", http.FileServer(http.Dir(".")))

	http.HandleFunc("/msg", func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err == nil {
			ch <- string(body)
		}

	})

	log.Fatal(http.ListenAndServe(":3047", nil))
}
