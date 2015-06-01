// Copyright 2015 David Wang. All rights reserved.
// Use of this source code is governed by MIT license.
// Please see LICENSE file

// WebSpark - lightweight web API front server
// Enables web services for Spark instances that can only make outbound http connections.
// Make these Spark instances available as web APIs
// WebSpark server supports large number of concurrent connections

// version 0.2

package main

import (
    "fmt"
    "net/http"
    "sync"
    "strings"
    "io/ioutil"
)

const ServerAddress=":8001"

// each instance is a spark provider for the named web API
type ProviderInstance struct {
	Name string // required for version 0.1: name is unique
	w *http.ResponseWriter // writer
	r *http.Request // reader
	channel chan string
}


type ClientInstance struct {
	Name, Path string 
	w *http.ResponseWriter // writer
	r *http.Request // reader 
	channel chan string
}


// WebSpark object
type WServer struct {
	providers map[string]*ProviderInstance
	clients map[string]*ClientInstance // web clients
	lock sync.RWMutex // concurrency R/W 
	counter int // client counter
}


func printNewConnection(r *http.Request, name string) {
	fmt.Println(name, r.RemoteAddr, r.URL)
}

func reportError(msg string) {
	fmt.Println(msg)
}

// register a new web API provider and URL
func (ws *WServer) AddProvider(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.URL.Query().Get("name"))
	if len(name)==0 {
		msg := "AddProvider: provider name required"
		reportError (msg)
		fmt.Fprintf(w, msg)
	} else {
		name = "/" + name
		printNewConnection(r, "Adding provider " + name)
		pi := ProviderInstance{name, &w, r, make(chan string)}
		ws.lock.Lock()
		ws.providers[name] = &pi
		ws.lock.Unlock()
		// ws.NewProvider(&pi)
		msg := <- pi.channel
		fmt.Fprintf(w, msg)
	}
}

// check for client waiting for provider
func (ws *WServer) NewProvider(pi *ProviderInstance) {
	ws.lock.Lock()
	for _, ci := range(ws.clients) {
		if ci.Path==pi.Name {
			// client for provider found
			ws.Pair(pi, ci)
		} 
	}
	ws.lock.Unlock()
}

// pair matched provider & client
// caller of this function needs to have write lock 
func (ws *WServer) Pair(pi *ProviderInstance, ci *ClientInstance) {
	fmt.Println("pairing", pi.Name, "with client", ci.Name)
	request := fmt.Sprintf("%s\n%s\n%s\n%s", ci.Name, ci.r.URL, ci.r.RemoteAddr, ci.r.Header)
	// fmt.Println(request)
	pi.channel<- request
	delete(ws.providers, ci.Path)
}

// register a new web client and routes request to provider as specified by URL
func (ws *WServer) AddClient(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if len(path)==0 {
		msg := "AddClient: URL path required"
		reportError (msg)
		fmt.Fprintf(w, msg)
	} else {
		printNewConnection(r, "Adding client for URL " + path)
		name := fmt.Sprintf("%d", ws.counter)
		ws.counter++
		ci := ClientInstance{name, path, &w, r, make(chan string)}
		ws.lock.Lock()
		ws.clients[name] = &ci
		ws.lock.Unlock()
		ws.NewClient(&ci)
		msg := <- ci.channel
		fmt.Fprintf(w, msg)
	}
}

// deliver provider result to paired client
func (ws *WServer) Respond(w http.ResponseWriter, r *http.Request) {
	clientname := strings.TrimSpace(r.URL.Query().Get("name"))
	if len(clientname)==0 {
		msg := "Respond: client name required"
		reportError (msg)
		fmt.Fprintf(w, msg)
	} else {
		printNewConnection(r, "Responding to " + clientname)
	    body, err := ioutil.ReadAll(r.Body)
	    if err != nil { 
	        reportError("cannot read provider response")
	    } else {
			ws.lock.Lock()
			ci, ok := ws.clients[clientname]
			if ok {
				ci.channel<- string(body)
				delete(ws.clients, clientname)
			} else {
				msg := fmt.Sprintf("no client found with id %s", clientname)
				reportError(msg)
			}			
			ws.lock.Unlock()
	    }
	}
}

// return a list of available API providers
func (ws *WServer) ListAPIs(w http.ResponseWriter, r *http.Request) {
	ws.lock.RLock()
	resp := ""
	for name, _ := range(ws.providers) {
		resp += name + "\n"
	}
	ws.lock.RUnlock()
	fmt.Fprintf(w, resp)
}

// handle new client request by URL path. Pass request to URL provider
func (ws *WServer) NewClient(ci *ClientInstance) {
	ws.lock.Lock()
	provider, ok := ws.providers[ci.Path]
	if ok {
		// provider is available for client
		ws.Pair(provider, ci)
	}
	ws.lock.Unlock()
}


func (ws *WServer) Favicon(w http.ResponseWriter, r *http.Request) {
	//ignore
}

func main() {
	//initialize server
	fmt.Println("Welcome to WebSpark. Serving on", ServerAddress)
	ws := new(WServer)
	ws.providers = make(map[string]*ProviderInstance)
	ws.clients = make(map[string]*ClientInstance)

	// add reserved URLs
    http.HandleFunc("/addapi", ws.AddProvider)
    http.HandleFunc("/respond", ws.Respond)
    http.HandleFunc("/listapis", ws.ListAPIs)
    http.HandleFunc("/favicon.ico", ws.Favicon)
    // add client URLs
    http.HandleFunc("/", ws.AddClient)
    http.ListenAndServe(ServerAddress, nil)
}

