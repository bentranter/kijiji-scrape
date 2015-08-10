# FoundIt

Get an email within five minutes of when the Kijiji item you want gets posted.

### Usage

Make sure you have Go installed on your system, and that your system is capable of handling the wildcard `*` operator at the filesystem level. After that, install dependencies:

```bash
$ go get github.com/yhat/scrape
$ go get golang.org/x/net/html
```

Once you've got those, you can just `$ go run`. Open your browser to port 3000 one it starts.