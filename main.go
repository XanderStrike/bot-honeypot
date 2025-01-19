package main

import (
    "log"
    "net/http"
)

func main() {
    // Serve static files from the current directory
    fs := http.FileServer(http.Dir("."))
    http.Handle("/", fs)

    log.Println("Starting server on :8080")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatal(err)
    }
}
