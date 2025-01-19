# 🤖 Bot Trap

A tiny honeypot that catches misbehaving bots and web scrapers in the act. Because if you're going to ignore robots.txt, you might as well end up in our hall of shame.

## What it does

1. Presents a perfectly normal webpage
2. ...with an invisible link that only bots would find
3. ...which is explicitly marked as off-limits in robots.txt
4. Keeps a running tally of naughty visitors
5. Also tracks vulnerability probing attempts, because why not?

## Running

```bash
go run main.go
```

Or with Docker:
```bash
docker build -t bot-trap .
docker run -p 8080:8080 bot-trap
```

Then sit back and watch the bots stumble in.

## But why?

Because sometimes you just want to watch the web crawl. 🕷️
