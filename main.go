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
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
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
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case "GET":
		err := tpl.ExecuteTemplate(w, "home", &page{"Home", "Welcome home.", nil})
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
			Email:    r.Form.Get("Email"),
		}
		matches := query.Scrape()

		err = tpl.ExecuteTemplate(w, "preview", &page{"Matches", "Showing all matches.", matches})
		if err != nil {
			log.Fatalln("Couldn't render page after POSTing: ", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		err = matches[0].Send(query.Email)
		if err != nil {
			log.Fatalln("Couldn't send email: ", err)
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
