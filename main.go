package main

import (
	"fmt"
	"net/http"
)

const AddForm = `
<form method="POST" action="/add">
URL: <input type="text" name="url">
<input type="submit" value="Add">
</form>
`

var store = NewURLStore("store.gob")

func main() {
	http.HandleFunc("/", redirect)
	http.HandleFunc("/add", add)
	if err := http.ListenAndServe(":8080", nil); err != nil {
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
	fmt.Fprintf(w, "http://localhost:8080/%s", key)
}
