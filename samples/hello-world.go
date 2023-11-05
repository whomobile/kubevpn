package main

import (
	"fmt"
	"net/http"
	"time"
)

const port = ":9080"

func handlerFunc(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%s from %s to %s\n", time.Now().Format(time.RFC3339), r.RemoteAddr, r.URL.Path)
	fmt.Println("-----------------------------------")

	r.ParseForm()
	if len(r.Form) > 0 {
		for key, values := range r.Form {
			for _, value := range values {
				fmt.Printf("Parameter %s = %s\n", key, value)
			}
		}
		fmt.Println("-----------------------------------")
	}

	for name, values := range r.Header {
		for _, value := range values {
			fmt.Printf("Header %s = %s\n", name, value)
		}
	}
	fmt.Println("===================================")

	fmt.Fprintf(w, "Hello, world!")
}

func main() {
	fmt.Printf("Server starting on port%s...\n", port)
	http.HandleFunc("/", handlerFunc)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Printf("Server failed: %s\n", err)
	}
}
