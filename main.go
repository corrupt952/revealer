package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
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
	IP      string      `json:"ip"`
	Headers http.Header `json:"headers"`
	Body    string      `json:"body"`
	Query   string      `json:"query"`
}

func (r *Response) String() string {
	json, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Sprintf("{\"error\": \"%s\"}", err)
	}
	return string(json)
}

var TRUSTED_PROXIES = []string{
	`^127\.0\.0\.1$`,
	`^::1$`,
	`^fc00:`,
	`^10\.`,
	`^172\.(1[6-9]|2[0-9]|3[0-1])\.`,
	`^192\.168\.`,
}
var TRUSTED_PROXY_REGEXP = regexp.MustCompile(strings.Join(TRUSTED_PROXIES, "|"))

func getClientIP(r *http.Request) string {
	remoteAddr := r.RemoteAddr
	clientIps := strings.Split(r.Header.Get("Client-Ip"), ",")
	forwardedIps := strings.Split(r.Header.Get("X-Forwarded-For"), ",")

	ips := clientIps[:0]
	ips = append(ips, forwardedIps...)
	ips = append(ips, remoteAddr)
	for _, ip := range ips {
		if ip != "" && !TRUSTED_PROXY_REGEXP.MatchString(ip) {
			return ip
		}
	}
	return remoteAddr
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
		IP:      getClientIP(r),
		Headers: getHeaders(r),
		Body:    string(body),
		Query:   r.URL.Query().Encode(),
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
