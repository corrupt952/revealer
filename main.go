package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"time"
)

/**
 * Middleware
 */
type Log struct {
	Time       string `json:"time"`
	RemoteAddr string `json:"remote_addr"`
	Method     string `json:"method"`
	Proto      string `json:"proto"`
	Host       string `json:"host"`
	Uri        string `json:"uri"`
	Query      string `json:"query"`
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := Log{
			Time:       time.Now().Format("2006-01-02 15:04:05"),
			RemoteAddr: r.RemoteAddr,
			Method:     r.Method,
			Host:       r.Host,
			Proto:      r.Proto,
			Uri:        r.RequestURI,
			Query:      r.URL.Query().Encode(),
		}
		json, err := json.Marshal(log)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fmt.Printf("%s\n", json)
		next.ServeHTTP(w, r)
	})
}

/**
 * Handlers
 */
type Response struct {
	RemoteAddr string      `json:"remote_addr"`
	Headers    http.Header `json:"headers"`
	Body       string      `json:"body"`
	Query      string      `json:"query"`
}

func (r *Response) String() string {
	json, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Sprintf("{\"error\": \"%s\"}", err)
	}
	return string(json)
}

func getHeaders(r *http.Request) http.Header {
	keys := make([]string, 0, len(r.Header))
	for k := range r.Header {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	headers := make(http.Header)
	for _, k := range keys {
		headers[k] = r.Header[k]
	}
	return headers
}

func buildResponse(r *http.Request) (Response, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return Response{}, err
	}
	response := Response{
		RemoteAddr: r.RemoteAddr,
		Headers:    getHeaders(r),
		Body:       string(body),
		Query:      r.URL.Query().Encode(),
	}
	return response, nil
}

func main() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response, err := buildResponse(r)
		if err != nil {
			fmt.Fprintf(w, "{\"error\": \"%s\"}\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "%s\n", response.String())
	})
	http.Handle("/", loggingMiddleware(handler))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Listening on port %s", port)
	http.ListenAndServe(":"+port, nil)
}
