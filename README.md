# Kijiji Scrape

Get an email within five minutes of when the Kijiji item you want gets posted. (Not really though)

### Usage

Make sure you have Go installed on your system, and that your system is capable of handling the wildcard `*` operator at the filesystem level. After that, install dependencies:

```bash
$ go get github.com/yhat/scrape
$ go get golang.org/x/net/html
$ go get gopkg.in/redis.v3
```

Once you've got those, you can just `$ go run`. Open your browser to port 3000 one it starts.

You'll also need Redis installed on your system. This repo uses the default Redis config, so if you've ran `$ brew install redis` or something similar, you can just `$ redis-server` and be on your way.

### License

Apache v2.0. See the license file for more info.

### Architecture

The idea is pretty simple:

1. Receive a request from the client.
1. Scrape a site (so far Kijiji is the only site implemented).
1. Look for one or more keywords.
1. Respond to the client immediately with the results (so that they feel like this web app is actually working).
1. Save the query in Redis and get it to puke back the ID back to you.
1. Start a goroutine that looks for updated keywords every five/ten minutes, and give it that ID.
1. Once a new result is found, email the user, and stop the routing. Ask them if they want to continue. looking or stop. The "continue looking" link is just `whatever.com/redisID?token=someCryptographicallyStrongToken&resume={boolean}`. From there,
    1. If the user wants to keep looking, just get the Redis ID from the URL, verify it's not forged using the token, get the query info from Redis, and start a new GoRoutine.
    1. If they don't want to keep looking, delete the query from Redis.

Why use Redis? Because I didn't think this through and I'm sure the server will crash (since I'm just going to be reckless and put this on a $5 DO droplet with no RAM lol), so at least if it fails I can restart without losing all the queries.