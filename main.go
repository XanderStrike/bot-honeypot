package main

import (
    "fmt"
    "html/template"
    "log"
    "net"
    "net/http"
    "strings"
    "sync"
    "time"
)

type Visitor struct {
    IP        string
    UserAgent string
    Timestamp time.Time
    Path      string    // Add path to track what they tried to access
    Type      string    // "secret" or "404"
}

// getRealIP extracts the real client IP from various proxy headers
func getRealIP(r *http.Request) string {
    // Try common proxy headers in order of preference
    headers := []string{
        "CF-Connecting-IP",  // Cloudflare
        "X-Forwarded-For",  // Common proxy header
        "X-Real-IP",        // Common proxy header
        "True-Client-IP",   // Cloudflare
        "Forwarded",        // RFC 7239
    }

    for _, header := range headers {
        if ip := r.Header.Get(header); ip != "" {
            // X-Forwarded-For can contain multiple IPs, take the first one
            if header == "X-Forwarded-For" {
                parts := strings.Split(ip, ",")
                if len(parts) > 0 {
                    ip = strings.TrimSpace(parts[0])
                }
            }
            // Validate the IP
            if parsedIP := net.ParseIP(ip); parsedIP != nil {
                return ip
            }
        }
    }

    // Fall back to RemoteAddr
    ip, _, err := net.SplitHostPort(r.RemoteAddr)
    if err != nil {
        return r.RemoteAddr // fallback to full RemoteAddr if parsing fails
    }
    return ip
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

func (vl *VisitorLog) Add(r *http.Request, visitorType string) {
    vl.mu.Lock()
    defer vl.mu.Unlock()

    vl.visitors[vl.current] = Visitor{
        IP:        getRealIP(r),
        UserAgent: r.UserAgent(),
        Timestamp: time.Now(),
        Path:      r.URL.Path,
        Type:      visitorType,
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

    // Handle robots.txt only route
    http.HandleFunc("/forbidden-scan", func(w http.ResponseWriter, r *http.Request) {
        visitorLog.Add(r, "forbidden")
        fmt.Fprintf(w, "🚨 Forbidden Area Detected! 🚨\n\nThis route only exists in robots.txt.\nYou're actively scanning forbidden content.\nYour IP and User-Agent have been logged. Naughty bot! 🤖")
    })

    // Handle javascript trap page
    http.HandleFunc("/javascript-trap", func(w http.ResponseWriter, r *http.Request) {
        visitorLog.Add(r, "javascript")
        fmt.Fprintf(w, "🕷️ Caught in the JavaScript Trap! 🕷️\n\nThis link was hidden in a JavaScript comment.\nOnly bots scraping with regex would find it.\nYour IP and User-Agent have been logged. Nice try, bot! 🤖")
    })

    // Handle secret page
    http.HandleFunc("/secret-page", func(w http.ResponseWriter, r *http.Request) {
        visitorLog.Add(r, "secret")
        fmt.Fprintf(w, "🚫 Gotcha! 🚫\n\nThis page was explicitly marked as off-limits in robots.txt.\nYour IP and User-Agent have been logged for posterity.\nMaybe try respecting robots.txt next time? 😉")
    })

    // Handle index page and catch all 404s
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path != "/" {
            if r.URL.Path != "/favicon.ico" && r.URL.Path != "/robots.txt" && r.URL.Path != "/secret-page" && r.URL.Path != "/javascript-trap" && r.URL.Path != "/forbidden-scan" {
                visitorLog.Add(r, "404")
            }
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
