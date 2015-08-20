// Package main implements a web scraper. It's basically a
// service that monitors sites like Kijiji or Etsy, and
// notifies you when a new ad matching the stuff you want
// gets posted. Here's my implmentation idea:
//
// 		- 2 pages: GET homepage and POST homepage
// 		- Add URL, email, keyword in field
// 		- start goroutine that looks until it finds match
// 		- email link once match is found, stop goroutine
//		- in email, ask 'look again'? If yes, restart
// 		  that goroutine, if no, end.
//
// Considerations:
//		- might not even need a DB?
// 		- might be able to implement fuzzy matches?
//		- follow links?
package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"time"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"gopkg.in/redis.v3"
)

var tpl = template.Must(template.ParseGlob("templates/*"))
var emailPassword, readErr = ioutil.ReadFile("./password.txt")

type page struct {
	Title   string
	Body    string
	Matches []*match
}

type query struct {
	SiteURL  string
	Keywords string
	Email    string
}

type match struct {
	Title       string
	Description string
	Link        string
	Price       string // Can be number, "swap/trade", or "please contact"
	Matched     bool
}

// Scrape scrapes a site for a keyword
func (q *query) Scrape() []*match {

	// Request the URL
	resp, err := http.Get(q.SiteURL)
	if err != nil {
		panic(err)
		log.Fatal("Couldn't GET ", q.SiteURL)
	}

	// Parse the contents of the URL
	root, err := html.Parse(resp.Body)
	if err != nil {
		panic(err)
		log.Fatal("Unable to parse response")
	}

	// Grab all the posts and print them
	posts := scrape.FindAll(root, scrape.ByClass("description"))
	matches := make([]*match, len(posts))
	for i, post := range posts {
		matches[i] = &match{
			Title:       scrape.Text(post.FirstChild.NextSibling),
			Description: scrape.Text(post),
			Link:        "http://kijiji.ca" + scrape.Attr(post.FirstChild.NextSibling, "href"),
			Price:       scrape.Text(post.NextSibling.NextSibling),
			Matched:     false,
		}
	}

	return matches
}

// Send emails the results of a successful match
func (m *match) Send(recipient string) error {

	to := []string{recipient}
	body := []byte(m.Description)

	username := "ben@boltmedia.ca"
	password := string(emailPassword)
	auth := smtp.PlainAuth("smtp.gmail.com:587", username, password, "smtp.gmail.com")

	return smtp.SendMail("smtp.gmail.com:587", auth, "ben@boltmedia.ca", to, body)
}

// HomeHandler handles the HTTP request
func HomeHandler(a *App, w http.ResponseWriter, r *http.Request) error {
	switch r.Method {

	case "GET":
		err := tpl.ExecuteTemplate(w, "home", &page{"Home", "Welcome home.", nil})
		if err != nil {
			log.Fatalln("Couldn't render home page template", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}
		return nil

	case "POST":
		// Get current time
		t := time.Now()

		err := r.ParseForm()
		if err != nil {
			panic(err)
			return err
		}

		fmt.Println("Time spent parsing form: ", time.Since(t))

		query := &query{
			SiteURL:  template.HTMLEscapeString(r.Form.Get("SiteURL")),
			Keywords: template.HTMLEscapeString(r.Form.Get("Keywords")),
			Email:    r.Form.Get("Email"),
		}
		matches := query.Scrape()

		fmt.Println("Time spent scraping form: ", time.Since(t))

		err = tpl.ExecuteTemplate(w, "preview", &page{"Matches", "Showing all matches.", matches})
		if err != nil {
			log.Fatalln("Couldn't render page after POSTing: ", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}

		fmt.Println("Time spent rendering template: ", time.Since(t))

		err = matches[0].Send(query.Email)
		if err != nil {
			log.Fatalln("Couldn't send email: ", err)
			return err
		}

		fmt.Println("Time spent sending email: ", time.Since(t))
		return nil
	}
	return nil
}

func initRedis() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	return client
}

func redisTestHandler(a *App, w http.ResponseWriter, r *http.Request) error {
	pong, err := a.DB.Ping().Result()
	if err != nil {
		return err
	}
	w.Write([]byte(pong))
	return nil
}

func main() {

	log.Println("Serving HTTP on port 3000")
	redisClient := initRedis()

	app := App{DB: redisClient}

	// Routes for handlers
	http.Handle("/", app.Handle(HomeHandler))
	http.Handle("/redis", app.Handle(redisTestHandler))

	// Static assets
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./css/"))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("./js/"))))

	// Start server
	http.ListenAndServe(":3000", nil)
}

// App stores gloabls & context vars
type App struct {
	DB *redis.Client
}

// Handler executes middleware and provides a ref to our
// global App struct
type Handler func(a *App, w http.ResponseWriter, r *http.Request) error

// Handle executes all our middleware. Each middleware
// function accepts an extra argument: the `a *app.App`,
// which is our global context variable
func (a *App) Handle(handlers ...Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, handler := range handlers {
			err := handler(a, w, r)
			if err != nil {
				// Lazily handle errors for now
				w.Write([]byte(err.Error()))
				return
			}
		}
	})
}
