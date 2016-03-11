package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"fmt"
	"time"
	"strings"
	"encoding/json"
	"os"
	"text/template"

	"github.com/gorilla/websocket"
)

type msg struct {
	Ts float64 `json:"t"`
	Data interface{} `json:"d"`
	Addr []string `json:"a"`
}

type hub struct {
	// Registered connections.
	connections map[*connection]bool

	// Inbound messages from the connections.
	broadcast chan []byte

	// Register requests from the connections.
	register chan *connection

	// Unregister requests from connections.
	unregister chan *connection

	// Data coming in.
	data chan msg;
}

func newHub() *hub {
	return &hub{
		broadcast:   make(chan []byte),
		register:    make(chan *connection),
		unregister:  make(chan *connection),
		connections: make(map[*connection]bool),
		data:        make(chan msg),
	}
}

func (h *hub) bcast (m []byte) {
	for c := range h.connections {
		select {
		case c.send <- m:
		default:
			delete(h.connections, c)
			close(c.send)
		}
	}
}

func marsh(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err == nil {
		return b;
	}
	return nil
}

func (h *hub) run() {
	for {
		select {

		case c := <-h.register:
			h.connections[c] = true

		case c := <-h.unregister:
			if _, ok := h.connections[c]; ok {
				delete(h.connections, c)
				close(c.send)
			}

		case m := <-h.broadcast:
			fmt.Printf("receive: %s\n", m)
			h.bcast(m)

		case s := <-h.data:
			t := time.Now()
			yy, mn, dy := t.Date()
			hh, mm, ss := t.Clock()
			as := strings.Join(s.Addr, "/")
			fn := fmt.Sprintf(
				"log-%04d-%02d-%02d.txt",
				yy, mn, dy)
			ln := fmt.Sprintf(
				"%02d:%02d:%02d: /%s %s\n",
				hh, mm, ss, as, marsh(s.Data))
			fmt.Print(ln)
			f, err := os.OpenFile(fn, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
			if err == nil {
				f.Write([]byte(ln))
				f.Close()
			} else {
				fmt.Printf("err=%v\n",err)
			}

			b := marsh(s)
			fmt.Printf("%v\n", string(b))
			h.bcast(b)
		}
	}
}

type connection struct {
	// The websocket connection.
	ws *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	// The hub.
	h *hub
}

func (c *connection) reader() {
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			break
		}
		c.h.broadcast <- message
	}
	c.ws.Close()
}

func (c *connection) writer() {
	for message := range c.send {
		err := c.ws.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			break
		}
	}
	c.ws.Close()
}

var upgrader = &websocket.Upgrader{
	ReadBufferSize: 1024, WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool { return true },
}

type wsHandler struct {
	h *hub
}

func (wsh wsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	c := &connection{send: make(chan []byte, 256), ws: ws, h: wsh.h}
	c.h.register <- c
	defer func() { c.h.unregister <- c }()
	go c.writer()
	c.reader()
}

func (h *hub) putdata (data []byte, addr []string) {
	var j interface{}
	err := json.Unmarshal(data, &j)
	if err == nil {
		t := time.Now()
		f := float64(t.UnixNano()) / 1.0e9;
		h.data <- msg{Ts: f, Data: j, Addr: addr}
	}
}

var homeTempl *template.Template

func homeHandler(c http.ResponseWriter, req *http.Request) {
	homeTempl.Execute(c, req.Host)
}

var (
	addr = flag.String("addr", ":3046", "http service address")
)

func main() {

	flag.Parse()

	h := newHub()
	go h.run()

	homeTempl = template.Must(template.ParseFiles("t.html"))
	http.HandleFunc("/msg", homeHandler)

	http.Handle("/", http.FileServer(http.Dir(".")))

	http.Handle("/msg/ws", wsHandler{h: h})

	http.HandleFunc("/msg/in", func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		//for k, v := range r.Header {
		//	fmt.Println(k, ": ", v)
		//}
		if err == nil {
			h.putdata(body, []string{})
		}

	})

	http.Handle("/msg/in/", http.StripPrefix("/msg/in/",http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err == nil {
			h.putdata(body, strings.Split(r.URL.Path, "/"))
		}

	})))

	log.Fatal(http.ListenAndServe(*addr, nil))
}
