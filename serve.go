package main

import (
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
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

func getScenesAmount(title string) (int, error) {
	files, err := filepath.Glob("data/" + title + "/" + "scene_*")
	if err != nil {
		return -1, err
	}
	if len(files) == 0 {
		return -1, errors.New(fmt.Sprintf("Empty project: %s contains no scene.", title))
	}
	return len(files), nil
}

func getCastsAmount(title string, scene string) (int, error) {
	files, err := filepath.Glob("data/" + title + "/" + scene + "/" + "asciicasts" + "/" + "file_*")
	if err != nil {
		log.Print(err)
		return -1, err
	}
	if len(files) == 0 {
		return -1, errors.New(fmt.Sprintf("Empty scene: %s contains no recording.", scene))
	}
	return len(files), nil
}

func loadProject(title string) (*Project, error) {
	scenesAmount, err := getScenesAmount(title)
	if err != nil {
		return nil, err
	}
	allScenes := make(map[int][]int)
	for i := 0; i < scenesAmount; i++ {
		sceneTitle := fmt.Sprintf("scene_%d", i+1)
		castsAmount, err := getCastsAmount(title, sceneTitle)
		if err != nil {
			log.Printf("The scene '%s' did not contain any recording", sceneTitle)
			continue
		} else if castsAmount > 0 { // Empty scenes are ignored.
			var allCasts []int
			for j := 0; j < castsAmount; j++ {
				allCasts = append(allCasts, j)
			}
			allScenes[i+1] = allCasts
		}
	}
	return &Project{Title: title, Scenes: allScenes}, nil
}

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

func loadHomePage() (*Page, error) {
	body, err := ioutil.ReadFile("static/homepage.html")
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
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
		log.Printf("Someone attempted to view %s but it hadn't been created yet.", title)
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
	fmt.Println("Body: " + body)
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
		// If no match
		if m == nil {
			// Return to home page
			http.Redirect(w, r, "/home.html", http.StatusFound)
			return
		}
		// runs the function (a handler)
		// m is the page's title
		fn(w, r, m[2])
	}
}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.Handle("/data/", http.StripPrefix("/data/", http.FileServer(http.Dir("data"))))
	http.Handle("/node_modules/", http.StripPrefix("/node_modules/", http.FileServer(http.Dir("node_modules"))))
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/post/", makeHandler(postHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
