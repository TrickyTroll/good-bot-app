package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
)

type Page struct {
	Title string // Title can be used to fetch contents.
	Body  []byte
}

type Project struct {
	Title  string
	Scenes map[int][]int // Maps scene # to amount of asciicasts in scene.
}

// panics if error in ParseFiles
var templates = template.Must(template.ParseFiles("tmpl/post.html", "tmpl/view.html"))

// panics if regexp does not compile

func (p *Page) save() error {
	filename := "data/" + p.Title + ".txt" // Using title as filename
	return ioutil.WriteFile(filename, p.Body, 0600)
}

/*
func (p *Page) saveYaml() error {
	filename := "data/" + p.Title + ".yaml"
	return ioutil.WriteFile(filename, p.Body, 0600)
}
*/

func loadPage(title string) (*Page, error) {
	filename := "data/" + title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil // Check second return value for errors.
	// Returns pointer to page filled
	// with correct info.
}

func getScenesAmount(title string) int {
	files, err := ioutil.ReadDir("data/" + title)
	if err != nil {
		return 0
	}
	return len(files)
}

func getCastsAmount(title string, scene string) int {
	files, err := ioutil.ReadDir("data/" + title + "/" + scene + "/" + "asciicasts")
	if err != nil {
		log.Print(err)
		return 0
	}
	return len(files)
}

func loadProject(title string) (*Project, error) {
	scenesAmount := getScenesAmount(title)
	allScenes := make(map[int][]int)
	for i := 0; i < scenesAmount; i++ {
		sceneTitle := fmt.Sprintf("scene_%d", i+1)
		castsAmount := getCastsAmount(title, sceneTitle)
		if castsAmount > 0 { // Empty scenes are ignored.
			var allCasts []int
			for j := 0; j < castsAmount; j++ {
				allCasts = append(allCasts, j)
			}
			allScenes[i+1] = allCasts
		}
	}
	return &Project{Title: title, Scenes: allScenes}, nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func renderProject(w http.ResponseWriter, tmpl string, p *Project) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadProject(title)
	if err != nil { // The post hasn't been created yet.
		http.Redirect(w, r, "/post/"+title, http.StatusFound)
		return
	}
	renderProject(w, "view", p) // Using the view template for now.
}

func postHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "post", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

var validPath = regexp.MustCompile("^/(post|save|view)/([a-zA-Z0-9]+)$")

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.Handle("/data/", http.StripPrefix("/data/", http.FileServer(http.Dir("data"))))
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/post/", makeHandler(postHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
