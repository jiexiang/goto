package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/rpc"
)

var (
	listenAddr = flag.String("http", ":8080", "http listen address")
	dataFile   = flag.String("file", "store.json", "data store")
	hostname   = flag.String("host", "localhost:8080", "http host name")
	masterAddr = flag.String("master", "", "RPC master address")
	rpcEnabled = flag.Bool("rpc", false, "enable RPC server")
)

var store Store

func main() {
	flag.Parse()
	if *masterAddr != "" {
		store = NewProxyStore(*masterAddr)
	} else {
		store = NewURLStore(*dataFile)
	}
	if *rpcEnabled { // the master is the rpc server:
		rpc.RegisterName("Store", store)
		rpc.HandleHTTP()
	}
	http.HandleFunc("/", redirect)
	http.HandleFunc("/add", add)
	if err := http.ListenAndServe(*listenAddr, nil); err != nil {
		panic(err)
	}
}

func redirect(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path[1:]
	if key == "" {
		http.NotFound(w, r)
		return
	}
	var url string
	if err := store.Get(&key, &url); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, url, http.StatusFound)
}

func add(w http.ResponseWriter, r *http.Request) {
	url := r.FormValue("url")
	// form 表单无数据，显示添加界面
	if url == "" {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, AddForm)
		return
	}
	var key string
	// form 表单有数据，保存
	if err := store.Put(&url, &key); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "http://%s/%s", *hostname, key)
}

const AddForm = `
<form method="POST" action="/add">
URL: <input type="text" name="url">
<input type="submit" value="Add">
</form>
`
