package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "backend" {
		fmt.Println("backend :3000")
		log.Fatal(http.ListenAndServe(":3000", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tp := r.Header.Values("Traceparent")
			fmt.Printf("traceparent count=%d values=%s\n", len(tp), strings.Join(tp, " | "))
			w.WriteHeader(200)
		})))
	}

	target, _ := url.Parse("http://localhost:3000")
	proxy := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(target)
			r.Out.Host = r.In.Host
		},
	}

	fmt.Println("proxy :80 -> :3000")
	log.Fatal(http.ListenAndServe(":80", proxy))
}
