package main

import (
    "fmt"
    "log"
    "net/http"
    "os"
    "time"
)

func logVisitor(r *http.Request) {
    ip := r.RemoteAddr
    userAgent := r.UserAgent()
    timestamp := time.Now().Format(time.RFC3339)
    
    logEntry := fmt.Sprintf("[%s] IP: %s, User-Agent: %s\n", timestamp, ip, userAgent)
    
    f, err := os.OpenFile("secret_visitors.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        log.Printf("Error opening log file: %v", err)
        return
    }
    defer f.Close()
    
    if _, err := f.WriteString(logEntry); err != nil {
        log.Printf("Error writing to log file: %v", err)
    }
}

func main() {
    // Handle robots.txt specifically
    http.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/plain")
        http.ServeFile(w, r, "robots.txt")
    })

    // Handle secret page
    http.HandleFunc("/secret-page", func(w http.ResponseWriter, r *http.Request) {
        logVisitor(r)
        fmt.Fprintf(w, "Welcome to the secret page!")
    })

    // Serve static files from the current directory
    fs := http.FileServer(http.Dir("."))
    http.Handle("/", fs)

    log.Println("Starting server on :8080")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatal(err)
    }
}
