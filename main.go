package main

import (
    "fmt"
    "html/template"
    "log"
    "net/http"
    "sync"
    "time"
)

type Visitor struct {
    IP        string
    UserAgent string
    Timestamp time.Time
}

type VisitorLog struct {
    visitors []Visitor
    current  int
    mu       sync.Mutex
}

func NewVisitorLog(size int) *VisitorLog {
    return &VisitorLog{
        visitors: make([]Visitor, size),
        current:  0,
    }
}

func (vl *VisitorLog) Add(r *http.Request) {
    vl.mu.Lock()
    defer vl.mu.Unlock()

    vl.visitors[vl.current] = Visitor{
        IP:        r.RemoteAddr,
        UserAgent: r.UserAgent(),
        Timestamp: time.Now(),
    }
    vl.current = (vl.current + 1) % len(vl.visitors)
}

func (vl *VisitorLog) GetAll() []Visitor {
    vl.mu.Lock()
    defer vl.mu.Unlock()

    // Create a sorted list starting from the oldest entry
    result := make([]Visitor, 0, len(vl.visitors))
    start := vl.current
    for i := 0; i < len(vl.visitors); i++ {
        idx := (start + i) % len(vl.visitors)
        if !vl.visitors[idx].Timestamp.IsZero() {
            result = append(result, vl.visitors[idx])
        }
    }
    return result
}

func main() {
    visitorLog := NewVisitorLog(100)
    
    // Handle robots.txt specifically
    http.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/plain")
        http.ServeFile(w, r, "robots.txt")
    })

    // Handle secret page
    http.HandleFunc("/secret-page", func(w http.ResponseWriter, r *http.Request) {
        visitorLog.Add(r)
        fmt.Fprintf(w, "Welcome to the secret page!")
    })

    // Handle index page
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path != "/" {
            http.NotFound(w, r)
            return
        }
        tmpl, err := template.ParseFiles("index.html")
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        data := struct {
            Visitors []Visitor
        }{
            Visitors: visitorLog.GetAll(),
        }
        tmpl.Execute(w, data)
    })

    log.Println("Starting server on :8080")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatal(err)
    }
}
