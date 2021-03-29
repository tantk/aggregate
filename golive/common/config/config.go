package config

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/tkanos/gonfig"
)

type key int

const (

	//LOG : Logging file
	requestIDKey key = 0
)

//Logging :
type Logging struct {
	Trace   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
}

//Configuration :
type Configuration struct {
	RESTport   string
	ServerLogs string
	ShopeeLogs string
	DbUser     string
	DbPW       string
	DbPort     string
	DbHost     string
	DbName     string
}

var (
	logger     *Logging
	once       sync.Once
	_, b, _, _ = runtime.Caller(0)

	// Root folder of this project
	Root = filepath.Join(filepath.Dir(b), "../../")
)

//GetConfig gets env var from a json file
func GetConfig(params ...string) Configuration {
	configuration := Configuration{}
	fileName := fmt.Sprintf(Root + "/common/config/conf.json")
	gonfig.GetConf(fileName, &configuration)
	return configuration
}

// CreateLogging : create custom logging obj for controllers.
//models logger can also be created seperately
func CreateLogging(filename string) *Logging {
	logFile, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	mw := io.MultiWriter(os.Stdout, logFile)
	logging := Logging{}

	trace := log.New(mw, "TRACE: ", log.Ldate|log.Ltime|log.Lshortfile)
	info := log.New(mw, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	warning := log.New(mw, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	error := log.New(io.MultiWriter(mw), "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	logging.Trace = trace
	logging.Info = info
	logging.Warning = warning
	logging.Error = error
	return &logging
}

//GetInstance is the a function to share the same logger across
// packages as long as the filename is the same
func GetInstance(s string) *Logging {
	once.Do(func() {
		logger = CreateLogging(s)
	})
	return logger
}

//Infologging :closure for http handler info logging
func (logger *Logging) Infologging() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				requestID, ok := r.Context().Value(requestIDKey).(string)
				if !ok {
					requestID = "unknown"
				}
				logger.Trace.Println(requestID, r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
			}()
			next.ServeHTTP(w, r)
		})
	}
}

//Tracing : logging middleware for gorilla mux
func Tracing() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-Id")
			if requestID == "" {
				requestID = fmt.Sprintf("%d", time.Now().UnixNano())
			}
			ctx := context.WithValue(r.Context(), requestIDKey, requestID)
			w.Header().Set("X-Request-Id", requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
