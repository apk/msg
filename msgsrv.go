// run-on-change msgsrv.go -- g build msgsrv.go -- ./msgsrv --addr 127.0.0.1:3046

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

	"github.com/gorilla/websocket"
)

type msg struct {
	Ts float64 `json:"t"`
	Data interface{} `json:"d"`
	Addr []string `json:"a"`
}

type act struct {
	Msg msg
	Action func(float64)
}

type evt struct {
	Action func(float64)
}

type request struct {
	Id interface{} `json:"$"`
	Cmd string `json:"!"`
	Val *json.RawMessage `json:"#"`
}

type response struct {
	Id interface{} `json:"$"`
	Ret interface{} `json:"?"`
	Val *json.RawMessage `json:"#"`
}

type select_param struct {
	Patterns [][]string `json:"patterns"`
}

type post_param struct {
	Addr []string `json:"a"`
	Data interface{} `json:"d"`
}

type post_reply struct {
	Ts float64 `json:"t"`
}

type cmsg struct {
	conn *connection
	msg []byte
}

type hub struct {
	// Registered connections.
	connections map[*connection]bool

	// Inbound messages from the connections.
	incoming chan cmsg

	// Register requests from the connections.
	register chan *connection

	// Unregister requests from connections.
	unregister chan *connection

	// Data coming in, plus an action.
	data chan act;

	// Events, only an action.
	evts chan evt;
}

func oops() {
}

func newHub() *hub {
	return &hub{
		incoming:    make(chan cmsg),
		register:    make(chan *connection),
		unregister:  make(chan *connection),
		connections: make(map[*connection]bool),
		data:        make(chan act, 512),           // TODO: Size
	}
}

func match_addr(pat [][]string, addr []string) bool {
	if len(pat) == 0 {
		return true
	}
	for _, p := range pat {
		for i, v := range addr {
			if i >= len(p) {
				return true
			}
			if v != p [i] {
				break
			}
		}
	}
	return false
}

func (h *hub) bcast (m []byte,addr []string) {
	for c := range h.connections {
		if match_addr(c.pattern,addr) {
			select {
			case c.send <- m:
			default:
				delete(h.connections, c)
				close(c.send)
			}
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

func write_file(fn string, data []byte) {
	f, err := os.OpenFile(fn, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
	if err == nil {
		defer f.Close();
		f.Write(data);
	} else {
		fmt.Printf("err=%v\n",err)
	}
}

func (h *hub) run() {

	events := []msg{}

	for {
		select {

		case c := <-h.register:
			h.connections[c] = true

		case c := <-h.unregister:
			if _, ok := h.connections[c]; ok {
				delete(h.connections, c)
				close(c.send)
			}

		case m := <-h.incoming:
			fmt.Printf("receive: %s\n", m)
			var req request
			err := json.Unmarshal(m.msg, &req)
			if err == nil {
				if req.Cmd == "select" {
					var sel select_param;
					err = json.Unmarshal(*req.Val, &sel)
					if err == nil {
						fmt.Printf("pattern: %v\n", sel.Patterns)
					}
					m.conn.pattern = sel.Patterns
				} else if req.Cmd == "post" {
					var post post_param;
					err = json.Unmarshal(*req.Val, &post)
					if err == nil {
						fmt.Printf("post: %v %v\n", post.Addr, post.Data)
						h.putdata (post.Data, post.Addr, func(f float64) {
							x, err := json.Marshal(post_reply{Ts: f})
							r := json.RawMessage(x)
							resp := response{Id: req.Id, Val: &r}
							// TODO: Dup code!
							b, err := json.Marshal(resp)
							if err == nil {
								m.conn.send <- b
							} else {
								_ = oops
							}
						})
					}
				} else {
					resp := response{Id: req.Id, Ret: false}
					b, err := json.Marshal(resp)
					if err == nil {
						m.conn.send <- b
					} else {
						_ = oops
					}
				}
			}
			// h.bcast(m)

		case s := <-h.evts:
			t := time.Now()
			f := float64(t.UnixNano()) / 1.0e9;
			s.Action(f)

		case s := <-h.data:
			t := time.Now()
			f := float64(t.UnixNano()) / 1.0e9;
			s.Msg.Ts = f
			yy, mn, dy := t.Date()
			hh, mm, ss := t.Clock()
			as := strings.Join(s.Msg.Addr, "/")
			fn := fmt.Sprintf(
				"log-%04d-%02d-%02d.txt",
				yy, mn, dy)
			ln := fmt.Sprintf(
				"%02d:%02d:%02d: /%s %s\n",
				hh, mm, ss, as, marsh(s.Msg.Data))
			write_file(fn, []byte(ln));

			b := marsh(s.Msg)
			h.bcast(b,s.Msg.Addr)

			events=append(events,s.Msg)

			secs := int(s.Msg.Ts) / 65536;
			fn = fmt.Sprintf("%02x", secs / 256);
			_ = os.Mkdir(fn, 0770);
			fn = fmt.Sprintf("%s/%02x", fn, secs % 256);
			write_file(fn, append(b, byte('\n')));
			s.Action(f)
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

	// The selections
	pattern [][]string
}

func (c *connection) reader() {
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			break
		}
		if err == nil {
			c.h.incoming <- cmsg{conn: c, msg: message}
		}
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

func (h *hub) putstrdata (data []byte, addr []string) {
	var j interface{}
	err := json.Unmarshal(data, &j)
	if err != nil {
		j = string(data)
	}
	h.putdata(j, addr, func(f float64) {})
}

func (h *hub) putevt (action func(float64)) {
	h.evts <- evt{Action: action}
}

func (h *hub) putdata (j interface{}, addr []string, action func(float64)) {
	h.data <- act{Msg: msg{Data: j, Addr: addr}, Action: action}
}

var (
	addr = flag.String("addr", ":3046", "http service address")
)

func main() {

	flag.Parse()

	h := newHub()
	go h.run()

	http.Handle("/", http.FileServer(http.Dir(".")))

	http.Handle("/msg/ws", wsHandler{h: h})

	http.HandleFunc("/msg/in", func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err == nil {
			h.putstrdata(body, []string{})
		}

	})

	http.Handle("/msg/in/", http.StripPrefix("/msg/in/",http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err == nil {
			h.putstrdata(body, strings.Split(r.URL.Path, "/"))
		}

	})))


        testfunc := func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		for k, v := range r.Header {
			fmt.Println(k, ": ", v)
		}
		if err == nil {
			fmt.Println("body: ", body)
		}
		fmt.Println("method: ", r.Method)
		fmt.Println("query: ", r.URL.RawQuery)

	};

	http.HandleFunc("/msg/test", testfunc);

	http.HandleFunc("/msg/test/", testfunc);

	log.Fatal(http.ListenAndServe(*addr, nil))
}
