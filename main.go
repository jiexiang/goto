package main

import (
	"flag"
	"fmt"
	"net/http"
)

var (
	listenAddr = flag.String("http", ":8080", "http listen address")
	dataFile   = flag.String("file", "store.gob", "data store")
	hostname   = flag.String("host", "localhost:8080", "http host name")
)

var store *URLStore

func main() {
	flag.Parse()
	store = NewURLStore(*dataFile)
	http.HandleFunc("/", redirect)
	http.HandleFunc("/add", add)
	if err := http.ListenAndServe(*listenAddr, nil); err != nil {
		panic(err)
	}
}

func redirect(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path[1:]
	url := store.Get(key)
	if url == "" {
		http.NotFound(w, r)
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
	// form 表单有数据，保存
	key := store.Put(url)
	fmt.Fprintf(w, "http://%s/%s", *hostname, key)
}

const AddForm = `
<form method="POST" action="/add">
URL: <input type="text" name="url">
<input type="submit" value="Add">
</form>
`
