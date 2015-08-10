// Package main implements `foundit.io`. Found It is a
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
	"log"
	"net/http"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
)

var tpl = template.Must(template.ParseGlob("templates/*"))

type page struct {
	Title string
	Body  string
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
			Description: scrape.Text(post.NextSibling.NextSibling),
			Link:        scrape.Text(post),
			Price:       scrape.Attr(post.FirstChild.NextSibling, "href"),
		}
		fmt.Printf("\033[32m%s\033[0m - \033[33m%s\033[0m\n%s\n\033[36mhttp://kijiji.ca%s\033[0m\n\n", scrape.Text(post.FirstChild.NextSibling), scrape.Text(post.NextSibling.NextSibling), scrape.Text(post), scrape.Attr(post.FirstChild.NextSibling, "href"))
	}

	return matches
}

// HomeHandler handles the HTTP request
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case "GET":
		err := tpl.ExecuteTemplate(w, "home", &page{"Home", "Welcome home."})
		if err != nil {
			log.Fatalln("Couldn't render home page template", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

	case "POST":
		err := r.ParseForm()
		if err != nil {
			panic(err)
		}

		query := &query{
			SiteURL:  template.HTMLEscapeString(r.Form.Get("SiteURL")),
			Keywords: template.HTMLEscapeString(r.Form.Get("Keywords")),
			Email:    template.HTMLEscapeString(r.Form.Get("Email")),
		}
		matches := query.Scrape()
		fmt.Println(matches[0])

		err = tpl.ExecuteTemplate(w, "home", &page{"Home - Post", "Posted to home."})
		if err != nil {
			log.Fatalln("Couldn't render home page after POSTing")
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func main() {

	log.Println("Serving HTTP on port 3000")

	// Routes for handlers
	http.HandleFunc("/", HomeHandler)

	// Static assets
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./css/"))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("./js/"))))

	// Start server
	http.ListenAndServe(":3000", nil)
}
