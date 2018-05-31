package main

import (
	"os"
	"fmt"
	"strconv"
	"net/http"
	"net"
	"syscall"
	"sync"
	"os/signal"
	"log"
	"io/ioutil"
	"strings"
	"time"
)

const AfDevKeyEnvVarName = "AF_DEV_KEY"
const AppodealAuthKeyName = "APPODEAL_AUTH_KEY"
const ListenPortEnvVarName = "AF_PROXY_PORT"
const HandlePattern = "/appsflyer_proxy/"
const AfBaseEndpoint = "https://api2.appsflyer.com/inappevent"
const HttpClientTimeout = time.Minute

func main() {
	log.Print("Load settings..")
	// get settings
	afDevKey, appodealAuthKey, listenPort, err := loadSettings()
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	log.Print("Settings loaded")

	// handle signals
	var signals = make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM)
	signal.Notify(signals, syscall.SIGQUIT)
	signal.Notify(signals, os.Interrupt)

	errLogger := log.New(os.Stdout, "[ERROR] ", log.LstdFlags)

	// wait group for handlers
	var wg sync.WaitGroup

	// Start listening port
	log.Printf("Start listening port %d", listenPort)
	server, err := net.Listen("tcp", fmt.Sprint(":", listenPort))
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	// Create handler
	mux := http.NewServeMux()
	mux.HandleFunc(HandlePattern, func(w http.ResponseWriter, r *http.Request) {
		wg.Add(1)
		defer wg.Done()

		// check header
		requestAuth := r.Header.Get("authentication")
		if requestAuth != appodealAuthKey {
		  panic("authentication header "+ requestAuth+" dont match")
		}

		routeParts := strings.Split(r.URL.Path, "/")
		if len(routeParts) != 3 {
			err = fmt.Errorf("Route is invalid: '%s'", r.URL.Path)
			errLogger.Print(err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		appBundleID := routeParts[2]

		// create HTTP POST request
		req, err := http.NewRequest("POST", genEndpoint(appBundleID), r.Body)
		if err != nil {
			panic(fmt.Errorf("Create request error - %v", err))
		}

		// set headers for HTTP request
		req.Header.Set("Content-Type", "application/json")
		// set authentication key
		log.Print("afDevKey"+afDevKey)
		req.Header.Set("authentication", afDevKey)

		// create HTTP client for AppsFlyer API
		afHTTPClient := http.Client{
			Transport: &http.Transport{
				MaxIdleConns:       1,
				IdleConnTimeout:    HttpClientTimeout,
				DisableCompression: true,
			},
		}
		// send request to AppsFlyer API
		resp, err := afHTTPClient.Do(req)
		if err != nil {
			panic(fmt.Errorf("Failed to send request - %v", err))
		}

		// read body
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			err = fmt.Errorf("Failed to read response body - %v", err)
			errLogger.Print(err)
			body = []byte(err.Error())
		}

    log.Printf("endpoint: ('%s')", genEndpoint(appBundleID))
		log.Printf("Response: (%d, '%s')", resp.StatusCode, string(body))

		if resp.StatusCode != http.StatusOK {
			err = fmt.Errorf("Response is not ok - %d. Body: '%s'", resp.StatusCode, string(body))
			errLogger.Print(err)
		}

		w.WriteHeader(resp.StatusCode)
		w.Write(body)
	})

	// Accept clients and process HTTP requests
	log.Print("Start server")
	go func() {
		if err := http.Serve(server, PanicProcessingMiddleware{Mux: mux, ErrLogger: errLogger}); err != nil {
			log.Fatal(err)
			signals <- syscall.SIGTERM
		}
	}()
	log.Print("Server started")

	<-signals      // wait for signals
	log.Println("going to shutdown...")
	server.Close() // stop HTTP server
	log.Println("server stopped")
	wg.Wait()      // wait for handlers
	log.Println("workers done")
}

func loadSettings() (afDevKey string, appodealAuthKey string, listenPort int, err error) {
	afDevKey = os.Getenv(AfDevKeyEnvVarName)
	if len(afDevKey) == 0 {
		err = fmt.Errorf(fmt.Sprint(AfDevKeyEnvVarName, " environment variable is not set!"))
		return
	}

	appodealAuthKey = os.Getenv(AppodealAuthKeyName)
  if len(appodealAuthKey) == 0 {
    err = fmt.Errorf(fmt.Sprint(AppodealAuthKeyName, " environment variable is not set!"))
    return
  }

	portEnv := os.Getenv(ListenPortEnvVarName)
	if len(portEnv) == 0 {
		err = fmt.Errorf(fmt.Sprint(ListenPortEnvVarName, " environment variable is not set!"))
		return
	}

	listenPort, err = strconv.Atoi(portEnv)
	if err != nil {
		err = fmt.Errorf(fmt.Sprint(ListenPortEnvVarName, " is not integer!"))
		return
	}
	if listenPort <= 0 {
		err = fmt.Errorf(fmt.Sprint(ListenPortEnvVarName, " must be > 0!"))
		return
	}
	return
}

func genEndpoint(appBundleID string) string {
	return fmt.Sprint(AfBaseEndpoint, "/", appBundleID)
}

// PanicProcessingMiddleware processing panics
type PanicProcessingMiddleware struct {
	Mux *http.ServeMux
	ErrLogger *log.Logger
}

// ServeHTTP serves HTTP connections
func (s PanicProcessingMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			err1 := fmt.Errorf("%+v", err)
			s.ErrLogger.Print(err1)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err1.Error()))
		}
	}()

	s.Mux.ServeHTTP(w, r)
}
