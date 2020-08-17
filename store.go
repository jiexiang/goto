package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
)

const saveQueueLength = 1000

type Store interface {
	Put(url, key *string) error
	Get(key, url *string) error
}

type URLStore struct {
	urls map[string]string
	mu   sync.RWMutex
	save chan record
}

type record struct {
	Key, URL string
}

func NewURLStore(filename string) *URLStore {
	s := &URLStore{
		urls: make(map[string]string),
		save: make(chan record, saveQueueLength),
	}
	if err := s.load(filename); err != nil {
		fmt.Println("load file error", err)
	}
	go s.saveLoop(filename)
	return s
}

func (s *URLStore) saveLoop(filename string) {
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("openFile error", err)
	}
	defer f.Close()
	e := json.NewEncoder(f)
	for {
		r := <-s.save
		if err := e.Encode(r); err != nil {
			fmt.Println("save error", err)
		}
	}
}

// 从 gob 文件中加载数据
func (s *URLStore) load(filename string) (err error) {
	f, err := os.Open(filename)
	if err != nil {
		return
	}
	defer f.Close()

	d := json.NewDecoder(f)
	for err == nil {
		var r record
		if err = d.Decode(&r); err == nil {
			s.Set(&r.Key, &r.URL)
		}
		if err == io.EOF {
			return nil
		}
	}
	return
}

func (s *URLStore) Get(key, url *string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if u, ok := s.urls[*key]; ok {
		*url = u
		return nil
	}
	return errors.New("key not found")
}

func (s *URLStore) Set(key, url *string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, present := s.urls[*key]; present {
		return errors.New("key already exists")
	}
	s.urls[*key] = *url
	return nil
}

func (s *URLStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.urls)
}

func (s *URLStore) Put(url, key *string) error {
	for {
		*key = genKey(s.Count())
		if err := s.Set(key, url); err == nil {
			break
		}
	}
	if s.save != nil {
		s.save <- record{
			Key: *key,
			URL: *url,
		}
	}
	return nil
}
