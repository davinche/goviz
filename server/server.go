package server

import (
	"html/template"
	"io/ioutil"
	"net"
	"net/http"
	"sync"

	"github.com/kardianos/osext"
	"golang.org/x/net/websocket"
)

var t *template.Template

func init() {
	binDir, err := osext.ExecutableFolder()
	if err != nil {
		return
	}
	t = template.Must(template.ParseFiles(binDir + "/static/index.html"))
}

func ListenAndServe(addr string) {
	mu := sync.Mutex{}
	clients := make(map[string][]*websocket.Conn)

	imagesMu := sync.Mutex{}
	images := make(map[string][]byte)

	http.HandleFunc("/connect", func(w http.ResponseWriter, r *http.Request) {
		id := r.FormValue("id")
		if id == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		handleWs := func(ws *websocket.Conn) {
			mu.Lock()
			clients[id] = append(clients[id], ws)
			mu.Unlock()

			img := images[id]
			if img != nil {
				websocket.Message.Send(ws, img)
			}

			for {
				var msg string
				err := websocket.Message.Receive(ws, &msg)
				if err != nil {
					mu.Lock()
					wsConnections := make([]*websocket.Conn, 0)
					for _, conn := range clients[id] {
						if conn != ws {
							wsConnections = append(wsConnections, conn)
						}
					}
					clients[id] = wsConnections
					mu.Unlock()
					return
				}
			}
		}
		websocket.Handler(handleWs).ServeHTTP(w, r)
	})

	http.HandleFunc("/generate", func(w http.ResponseWriter, r *http.Request) {
		id := r.FormValue("id")
		if id == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		defer r.Body.Close()
		img, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		mu.Lock()
		c := clients[id]
		mu.Unlock()

		if c != nil {
			for _, ws := range c {
				websocket.Message.Send(ws, img)
			}
		}

		imagesMu.Lock()
		images[id] = img
		imagesMu.Unlock()
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		id := r.FormValue("id")
		if id == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		host, port, err := net.SplitHostPort(r.Host)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		s := struct {
			Host string
			Port string
			ID   string
		}{
			Host: host,
			Port: port,
			ID:   id,
		}
		t.ExecuteTemplate(w, "index.html", s)
	})
	http.ListenAndServe(addr, nil)
}
