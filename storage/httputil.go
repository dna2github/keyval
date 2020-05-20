package storage

import (
	"net/http"
	"fmt"
	"strings"
	"io/ioutil"
)

const (
	apiPrefix    = "/api/v1/"
	maxValueSize = 1024 * 1024 * 4
)

var defaultStorage = StorageFilesystem{}

func StorageHandler(w http.ResponseWriter, r *http.Request) {
	uri := r.URL.RequestURI()
	if !strings.HasPrefix(uri, apiPrefix) {
		http.NotFound(w, r)
		return
	}
	uri = strings.TrimPrefix(uri, apiPrefix)
	parts := strings.Split(uri, "/")
	// reserved to be used
	// _ = parts[0]
	if len(parts) <= 2 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	uri = fmt.Sprintf("%s://%s", parts[1], strings.Join(parts[2:], "/"))
	switch r.Method {
	case "GET":
		getValue(uri, w, r)
		return
	case "POST":
		putKeyValue(uri, w, r)
		return
	case "PUT":
		putKeyValue(uri, w, r)
		return
	case "DELETE":
		delKey(uri, w, r)
		return
	}
	// PATCH, OPTION, ...
	w.Header().Add("Content-Type", "text/plain")
	fmt.Fprint(w, uri)
}

func getValue(url string, w http.ResponseWriter, r *http.Request) {
	key := defaultStorage.UrlToKey(url)
	b, ok := defaultStorage.Get(key)
	if !ok {
		http.NotFound(w, r)
		return
	}
	w.Header().Add("Content-Type", "text/plain")
	w.Write(b)
}

func putKeyValue(url string, w http.ResponseWriter, r *http.Request) {
	key := defaultStorage.UrlToKey(url)
	r.Body = http.MaxBytesReader(w, r.Body, maxValueSize)	
	value, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	ok := defaultStorage.Put(key, value)
	if !ok {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	w.Header().Add("Content-Type", "text/plain")
	w.Write([]byte(url))
}

func delKey(url string, w http.ResponseWriter, r *http.Request) {
	key := defaultStorage.UrlToKey(url)
	_, ok := defaultStorage.Get(key)
	if !ok {
		http.NotFound(w, r)
		return
	}
	ok = defaultStorage.Del(key)
	if !ok {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "text/plain")
	w.Write([]byte(url))
}