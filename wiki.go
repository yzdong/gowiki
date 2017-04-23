package main

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
)

var Error *log.Logger
var dataDirName = "data"

func Init(errorHandle io.Writer) {
	Error = log.New(errorHandle, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

type Page struct {
	Title    string
	Location string
	Body     []byte
}

type PageInterface interface {
	save(string) error
	addLinks(string)
}

func (p *Page) save(s string) error {
	filename := s + "/" + p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func (p *Page) addLinks(keyword string) {
	pattern := fmt.Sprintf("[^>/]%s[^<>]|^%s$", keyword, keyword)
	re := regexp.MustCompile(pattern)
	repl := []byte(`<a href="/view/` + keyword + `">` + keyword + `</a>`)
	after := re.ReplaceAll(p.Body, repl)
	p.Body = after
}

type Pages struct {
	All []PageInterface
}

type PagesInterface interface {
	addLinksToPages(string) []error
}

func (ps *Pages) addLinksToPages(keyword string) []error {
	errs := make([]error, 0, len(ps.All))
	for _, value := range ps.All {
		value.addLinks(keyword)
		err := value.save(dataDirName)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func loadPage(title string) (*Page, error) {
	filename := dataDirName + "/" + title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Location: filename, Body: body}, nil
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string, pages PagesInterface) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string, pages PagesInterface) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string, pages PagesInterface) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save(dataDirName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	errs := pages.addLinksToPages(title)
	if len(errs) != 0 {
		for _, val := range errs {
			Error.Println(val.Error())
		}
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

var templates = template.Must(template.ParseFiles("tmpl/edit.html", "tmpl/view.html", "index.html"))

type EscapedPage struct {
	Title string
	Body  template.HTML
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	ep := &EscapedPage{Title: p.Title, Body: template.HTML(p.Body)}
	err := templates.ExecuteTemplate(w, tmpl+".html", ep)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func initPages(dir string) *Pages {
	re := regexp.MustCompile(".txt")
	pages := &Pages{All: make([]PageInterface, 0, 10)}
	files, _ := ioutil.ReadDir(dir)
	for _, val := range files {
		title := re.ReplaceAllString(val.Name(), "")
		page, _ := loadPage(title)
		pages.All = append(pages.All, page)
	}
	return pages
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string, PagesInterface), ps PagesInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2], ps)
	}
}

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-0]+)$")

func main() {
	Init(os.Stderr)
	pages := initPages(dataDirName)
	http.HandleFunc("/view/", makeHandler(viewHandler, pages))
	http.HandleFunc("/edit/", makeHandler(editHandler, pages))
	http.HandleFunc("/save/", makeHandler(saveHandler, pages))
	http.HandleFunc("/", defaultHandler)
	http.ListenAndServe(":8080", nil)
}
