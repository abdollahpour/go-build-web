package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	"gopkg.in/russross/blackfriday.v2"
)

// Convert markown text to HTML
func markdownFunc(t string) template.HTML {
	html := blackfriday.Run([]byte(t))
	return template.HTML(html)
}

// Convert markdown file to HTML
func markdownFileFunc(path string) template.HTML {
	absPath, err := filepath.Abs(path)
	var file *os.File
	if absPath == path {
		file, err = os.Open(absPath)
	} else {
		file, err = os.Open(filepath.Join("static", path))
	}
	if err != nil {
		os.Stderr.WriteString("Error to access markdown file: " + path + "\n")
		return template.HTML("")
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		os.Stderr.WriteString("Error to read the markdown file\n")
		return template.HTML("")
	}

	html := blackfriday.Run(bytes)
	return template.HTML(html)
}

func buildAll(source string, destination string) error {
	// get properties of source dir
	info, err := os.Stat(source)
	if err != nil {
		return err
	}

	if info.IsDir() {
		// create dest dir
		err = os.MkdirAll(destination, info.Mode())
		if err != nil {
			return fmt.Errorf("Error to create a directory: %s", destination)
		}

		dir, _ := os.Open(source)
		files, err := dir.Readdir(-1)
		if err != nil {
			return fmt.Errorf("Error to list the files: %s", destination)
		}

		for _, f := range files {
			err = buildAll(source+"/"+f.Name(), destination+"/"+f.Name())
			if err != nil {
				return nil
			}
		}
	} else {
		if source == "static/layout.html" || strings.HasSuffix(source, "/index.json") || strings.HasSuffix(source, ".md") {
			return nil
		}

		if strings.HasSuffix(source, "/index.html") {
			dest, err := os.Create(destination)
			if err != nil {
				return fmt.Errorf("Error to create: %s", destination)
			}
			defer dest.Close()

			err = build(source, dest, nil)
			if err != nil {
				return err
			}
			return nil
		}

		// perform copy
		// TODO: Check if the destination file is newser, don't copy it
		if !info.Mode().IsRegular() {
			return fmt.Errorf("%s is not a regular file", source)
		}

		src, err := os.Open(source)
		if err != nil {
			return fmt.Errorf("Error to open: %s", source)
		}
		defer src.Close()

		dest, err := os.Create(destination)
		if err != nil {
			return fmt.Errorf("Error to create: %s", destination)
		}
		defer dest.Close()
		_, err = io.Copy(dest, src)
		if err != nil {
			return fmt.Errorf("Error to copy '%s' from '%s'", source, destination)
		}
		fmt.Printf("'%s' copied\n", source)
	}

	return nil
}

func build(html string, w io.Writer, httpWriter http.ResponseWriter) error {
	// TODO: How about if the file named *.JSON?
	// Remove static from begin
	path := filepath.Dir(html[6:]) + "/"
	jsonFile := html[:len(html)-5] + ".json"

	var data map[string]interface{}
	jsonData, err := ioutil.ReadFile(jsonFile)

	if err == nil {
		err = json.Unmarshal([]byte(jsonData), &data)
		if err != nil {
			return fmt.Errorf("Error to parse the JSON meta data file: %s", jsonFile)
		}
	}

	tmpl := template.Must(template.New("base").Funcs(funcMap(path)).ParseFiles(templates(html)...))
	err = tmpl.ExecuteTemplate(w, "layout", data)
	if err != nil {
		return fmt.Errorf("Error to build the template: %s", html)
	}
	fmt.Printf("Template compiled: %s\n", html)

	return nil
}

// We provide some extra functionality in our templates files to help generate HTML files easier (ex using Markdown)
func funcMap(path string) map[string]interface{} {
	fmt.Println(path)
	return template.FuncMap{
		"StringsJoin": strings.Join,
		"Markdown":    markdownFunc,
		"PathStarts": func(arg string) string {
			if strings.HasPrefix(path, arg) {
				return arg
			}
			return ""
		},
		"PathEquals": func(arg string) string {
			if path == arg {
				return arg
			}
			return ""
		},
		"MarkdownFile": func(arg string) template.HTML {
			return markdownFileFunc(path + arg)
		},
	}
}

func templates(html string) []string {
	var layouts []string
	layouts = append(layouts, html)
	for html != "static" {
		html = filepath.Dir(html)
		layout := html + "/layout.html"
		if _, err := os.Stat(layout); !os.IsNotExist(err) {
			layouts = append(layouts, layout)
		}

	}
	return layouts
}

func main() {
	// Run predefined commands, clean & build
	if len(os.Args) > 1 {
		for _, arg := range os.Args {
			if strings.EqualFold(arg, "clean") {
				// Clean the generated files
				os.RemoveAll("build/")
			} else if strings.EqualFold(arg, "build") {
				// Build the static website
				fmt.Println("Builing the project ...")
				err := buildAll("static", "build")
				if err != nil {
					panic(err)
				}
			}
		}
		return
	}

	r := mux.NewRouter()

	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		path := req.URL.Path
		index := "static" + path + "index"
		html := index + ".html"
		meta := index + ".json"

		// Load the json file using default params
		var data map[string]interface{}
		file, err := ioutil.ReadFile(meta)
		if err == nil {
			_ = json.Unmarshal([]byte(file), &data)
		}

		// TODO: We disable the Links for now. The idea is for static generate version, we monitor loaded resource and generate links
		// for _, preload := range data.Preloads {
		//	_as := preloads[filepath.Ext(preload)]
		//	if len(_as) > 0 {
		//		w.Header().Add("Link", "<"+preload+">; as="+_as+"; rel=preload")
		//	}
		//}
		tmpl := template.Must(template.New("base").Funcs(funcMap(path)).ParseFiles(templates(html)...))
		tmpl.ExecuteTemplate(w, "layout", data)
	}).MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) bool {
		if strings.HasSuffix(r.URL.Path, "/") {
			_, err := os.Stat("static/" + r.URL.Path + "index.html")
			return !os.IsNotExist(err)
		}
		return false
	})

	// ============================= Handle your rest API here ==============================
	// If you want to have some rest APIs and turn your static website into a fully dynamic
	// Website, you can easily add your handlers here and consume that rest API in your pages
	// ======================================================================================

	fs := http.FileServer(http.Dir("static"))
	r.PathPrefix("/").Handler(http.StripPrefix("/", fs))

	http.Handle("/", r)
	http.ListenAndServe(":3000", nil)
}
