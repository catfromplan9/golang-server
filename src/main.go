package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var html_path = "./html"
var url_root = ""
var www_path = "/var/www/example.net/html"
var site_title = "Example.net"

var static_login []byte
var static_upload []byte

func headers(w http.ResponseWriter, req *http.Request) {
	for name, headers := range req.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
	fmt.Fprintln(w, "Cookies:")
	for _, cookie := range req.Cookies() {
		fmt.Println("Found a cookie named:", cookie.Name)
	}
}

func handler_stats(w http.ResponseWriter, req *http.Request) {
	var out []byte
	var err error

	out, err = exec.Command("df").Output()
	if err != nil {
		handler_error(w, req, 500)
	}
	fmt.Fprintln(w, "# df")
	fmt.Fprint(w, string(out))

	out, err = exec.Command("free").Output()
	if err != nil {
		handler_error(w, req, 500)
	}
	fmt.Fprintln(w, "# free")
	fmt.Fprint(w, string(out))

	out, err = exec.Command("uptime").Output()
	if err != nil {
		handler_error(w, req, 500)
	}
	fmt.Fprintln(w, "# uptime")
	fmt.Fprint(w, string(out))

}

func handler_error(w http.ResponseWriter, req *http.Request, code int) {
	statustext := http.StatusText(code)
	http.Error(w, statustext, code)
}

func handler_teapot(w http.ResponseWriter, req *http.Request) {
	handler_error(w, req, http.StatusTeapot)
}

type TemplateServer struct {
	cache map[string]*template.Template
	mtx   sync.Mutex
}

func (ts *TemplateServer) clear(w http.ResponseWriter, req *http.Request) {
	ts.cache = make(map[string]*template.Template)
	fmt.Fprintln(w, "Cache cleared")
}

func (ts *TemplateServer) create(fp string) error {
	ts.mtx.Lock()
	defer ts.mtx.Unlock()
	var err error
	if ts.cache[fp] == nil {
		var tmpl *template.Template

		if filepath.Ext(fp) == ".md" {
			/* Markdown file, convert it */
			log.Println("Converting Markdown to HTML " + fp)

			cmd := exec.Command("node", "./markdown/convert.mjs", fp)
			err := cmd.Run()
			if err != nil {
				return err
			}

			tmpl = template.New(fp)

			f, err := os.ReadFile("buffer")
			_, err = tmpl.New(filepath.Base(fp)).Parse(string(f))
			if err != nil {
				return err
			}

			_, err = tmpl.ParseFiles("./html/template.html")
			if err != nil {
				return err
			}

		} else {
			/* Normal HTML file */
			tmpl, err = template.ParseFiles("./html/template.html", fp)
			if err != nil {
				return err
			}
		}

		ts.cache[fp] = tmpl
		log.Println("Cached a template for " + fp)
	}
	return nil
}

func (ts *TemplateServer) serve(w http.ResponseWriter, r *http.Request) {

	fp := filepath.Join(html_path, filepath.Clean(r.URL.Path))

	/* Redirect index.html to root */
	if base := filepath.Base(fp); base == "index.html" {
		log.Println("Redirecting index.html to root")
		http.Redirect(w, r, r.URL.Path[0:len(r.URL.Path)-len(base)], http.StatusSeeOther)
		return
	}

	info, err := os.Stat(fp)
	if err != nil {
		handler_error(w, r, http.StatusNotFound)
		return
	}

	if info.IsDir() {
		/* Redirect folder to folder/ */
		if r.URL.Path[len(r.URL.Path)-1:] != "/" {
			http.Redirect(w, r, r.URL.Path+"/", http.StatusSeeOther)
			return
		}

		fp = filepath.Join(fp, "index.html")
	}

	if filepath.Ext(fp) == ".md" && err == nil {
		/* Serve raw MD files */
		dat, err := os.ReadFile(fp)
		if err != nil {
			handler_error(w, r, http.StatusInternalServerError)
			return
		}
		txt := string(dat)
		fmt.Fprint(w, txt)
		return
	}

	if filepath.Ext(fp) != ".html" {
		http.ServeFile(w, r, fp)
		return
	}

	/* If *.html does not exist, check for *.md */
	if _, err := os.Stat(fp); err != nil {
		var ext = filepath.Ext(fp)
		var fpmd = fp[0:len(fp)-len(ext)] + ".md"
		if _, err := os.Stat(fpmd); err == nil {
			fp = fpmd
		}
	}

	info, err = os.Stat(fp)
	if err != nil {
		handler_error(w, r, http.StatusNotFound)
		return
	}
	log.Println("Serving " + fp + " for " + r.URL.Path)

	err = ts.create(fp)
	if err != nil {
		fmt.Println(fp)
		fmt.Println(err)
		handler_error(w, r, http.StatusInternalServerError)
		return
	}

	tmpl := ts.cache[fp]

	err = tmpl.ExecuteTemplate(w, filepath.Base(fp), nil)

	if err != nil {
		fmt.Println(err)
		handler_error(w, r, http.StatusInternalServerError)
		return
	}
}

func main() {
	log.Println("Run")
	rand.Seed(time.Now().UTC().UnixNano())

	var err error
	static_login, err = ioutil.ReadFile(html_path + "/login.html")

	if err != nil {
		log.Fatal("File read failed")
		os.Exit(1)
	}

	ts := TemplateServer{
		cache: make(map[string]*template.Template),
	}

	http.HandleFunc("/_stats", handler_stats)
	http.HandleFunc("/_clear", ts.clear)

	http.HandleFunc("/teapot", handler_teapot)
	http.HandleFunc("/_headers", headers)
	http.HandleFunc("/", ts.serve)

	if err == nil {
		err = db_init()
	}
	if err == nil {
		err = accounts_init()
	}

	if err != nil {
		fmt.Println(err)
		log.Fatal("Error")
		os.Exit(1)
	}

	log.Println("Attempting to listen :5001")

	log.Fatal(http.ListenAndServe(":5001", nil))

}
