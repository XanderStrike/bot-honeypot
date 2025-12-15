package main

import (
    "encoding/json"
    "fmt"
    "html/template"
    "log"
    "net"
    "net/http"
    "os"
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

const visitorLogFile = "visitors.json"

type VisitorLog struct {
    mu       sync.Mutex
}

func NewVisitorLog() *VisitorLog {
    return &VisitorLog{}
}

func (vl *VisitorLog) loadVisitors() ([]Visitor, error) {
    vl.mu.Lock()
    defer vl.mu.Unlock()

    // Try to read existing file
    data, err := os.ReadFile(visitorLogFile)
    if err != nil {
        if os.IsNotExist(err) {
            // File doesn't exist, return empty slice
            return []Visitor{}, nil
        }
        return nil, err
    }

    var visitors []Visitor
    if err := json.Unmarshal(data, &visitors); err != nil {
        return nil, err
    }

    return visitors, nil
}

func (vl *VisitorLog) saveVisitors(visitors []Visitor) error {
    vl.mu.Lock()
    defer vl.mu.Unlock()

    data, err := json.MarshalIndent(visitors, "", "  ")
    if err != nil {
        return err
    }

    return os.WriteFile(visitorLogFile, data, 0644)
}

func (vl *VisitorLog) Add(r *http.Request, visitorType string) {
    visitors, err := vl.loadVisitors()
    if err != nil {
        log.Printf("Error loading visitors: %v", err)
        return
    }

    // Add new visitor
    visitors = append(visitors, Visitor{
        IP:        getRealIP(r),
        UserAgent: r.UserAgent(),
        Timestamp: time.Now(),
        Path:      r.URL.Path,
        Type:      visitorType,
    })

    // Save updated visitors
    if err := vl.saveVisitors(visitors); err != nil {
        log.Printf("Error saving visitors: %v", err)
    }
}

func (vl *VisitorLog) GetAll() []Visitor {
    visitors, err := vl.loadVisitors()
    if err != nil {
        log.Printf("Error loading visitors: %v", err)
        return []Visitor{}
    }
    return visitors
}

func main() {
    visitorLog := NewVisitorLog()
    
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
