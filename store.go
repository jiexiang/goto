package main

import (
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"sync"
)

type URLStore struct {
	urls map[string]string
	mu   sync.RWMutex
	file *os.File
}

type record struct {
	Key, URL string
}

func NewURLStore(filename string) *URLStore {
	s := &URLStore{
		urls: make(map[string]string),
	}
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("openFile error", err)
	}
	s.file = f
	if err := s.load(); err != nil {
		fmt.Println("load error", err)
	}
	return s
}

// 从 gob 文件中加载数据
func (s *URLStore) load() (err error) {
	if _, err = s.file.Seek(0, 0); err != nil {
		return
	}
	d := gob.NewDecoder(s.file)
	for err == nil {
		var r record
		if err = d.Decode(&r); err == nil {
			s.Set(r.Key, r.URL)
		}
		if err == io.EOF {
			return nil
		}
	}
	return
}

func (s *URLStore) Get(key string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.urls[key]
}

func (s *URLStore) Set(key, url string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, present := s.urls[key]; present {
		return false
	}
	s.urls[key] = url
	return true
}

func (s *URLStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.urls)
}

func (s *URLStore) Put(url string) string {
	for {
		key := genKey(s.Count())
		if ok := s.Set(key, url); ok {
			if err := s.save(key, url); err != nil {
				fmt.Println("save error", err)
			}
			return key
		}
	}
	return ""
}

// 存储到文件
func (s *URLStore) save(key, url string) error {
	e := gob.NewEncoder(s.file)
	return e.Encode(record{key, url})
}
