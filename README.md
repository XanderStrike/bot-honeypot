# Bot Honeypot

A tiny honeypot that catches misbehaving bots and web scrapers in the act.

## What it does

Presents a set of traps for automated web agents:

* An invisible link to a route disallowed in robots.txt
* A route only present in a javascript comment
* A route only present in the robots.txt, where it's disallowed 
* Logging for 404s, showing vulnerability probers

## Running

```bash
go run main.go
```

Or with Docker:
```bash
docker run -p 8080:8080 xanderstrike/bot-honeypot
```

Then visit http://localhost:8080 to see caught bots in real-time.

Hint: Put this behind a domain and then get a letsencrypt cert for it.

