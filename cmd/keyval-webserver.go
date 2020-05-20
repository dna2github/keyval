package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"net/http"
	"github.com/dna2zodiac/keyval/storage"
)

type ServerConfig struct {
	debug            bool
	serverStaticDir  string
	serverHttpsCADir string
	serverHttpsCertFilename string
	serverHttpsKeyFilename  string
	serverListen     string
}

func parseEnv() (s ServerConfig) {
	s.debug = len(os.Getenv("KEYVAL_DEBUG")) > 0
	s.serverStaticDir = os.Getenv("KEYVAL_STATIC_DIR")
	s.serverHttpsCADir = os.Getenv("KEYVAL_HTTPS_CA_DIR")

	serverHost := os.Getenv("KEYVAL_HOST")
	serverPort, _ := strconv.Atoi(os.Getenv("KEYVAL_PORT"))
	if serverPort <= 0 {
		serverPort = 8080
	}
	s.serverListen = fmt.Sprintf("%s:%d", serverHost, serverPort)


	if len(s.serverHttpsCADir) > 0 {
		s.serverHttpsCertFilename = path.Join(s.serverHttpsCADir, "ca.pem")
		s.serverHttpsKeyFilename = path.Join(s.serverHttpsCADir, "ca.key")
		fileInfo, err := os.Stat(s.serverHttpsCertFilename)
		if os.IsNotExist(err) || fileInfo.IsDir() {
			s.serverHttpsCertFilename = ""
			s.serverHttpsKeyFilename = ""
		}
		fileInfo, err = os.Stat(s.serverHttpsKeyFilename)
		if os.IsNotExist(err) || fileInfo.IsDir() {
			s.serverHttpsCertFilename = ""
			s.serverHttpsKeyFilename = ""
		}
	}

	if len(s.serverStaticDir) > 0 {
		fileInfo, err := os.Stat(s.serverStaticDir)
		if os.IsNotExist(err) || !fileInfo.IsDir() {
			s.serverStaticDir = ""
		}
	}

	return
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain")
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func buildServer(mux *http.ServeMux) {
	mux.HandleFunc("/hello/", helloHandler)
	mux.HandleFunc("/api/v1/", storage.StorageHandler)
}

func main() {
	config := parseEnv()
	mux := http.NewServeMux()
	if config.serverStaticDir != "" {
		fileServer := http.FileServer(http.Dir(config.serverStaticDir))
		mux.Handle("/static/", http.StripPrefix("/static", fileServer))
	}
	buildServer(mux)

	fmt.Println("keyval is running at", config.serverListen, "...")
	if len(config.serverHttpsCertFilename) > 0 {
		log.Fatal(http.ListenAndServeTLS(config.serverListen, config.serverHttpsCertFilename, config.serverHttpsKeyFilename, mux))
	} else {
		log.Fatal(http.ListenAndServe(config.serverListen, mux))
	}
}