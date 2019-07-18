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
	"regexp"
	"strings"

	"github.com/gorilla/mux"
	"gitlab.com/golang-commonmark/markdown"
)

// Page meta data
type page struct {
	Templates   []string
	Preloads    []string
	Language    string
	Description string
	Keywords    []string
	Author      string
	Image       string
	Title       string
	URL         string
}

// Preload format, we just support these extention, anything else will ignore
var preloads = map[string]string{
	".js":   "script",
	".css":  "style",
	".png":  "image",
	".jpg":  "image",
	".jpeg": "image",
	".gif":  "image",
	".svg":  "image",
}

// Convert markown text to HTML
func markdownFunc(t string) template.HTML {
	md := markdown.New(markdown.XHTMLOutput(true))
	html := md.RenderToString([]byte(t))
	return template.HTML(html)
}

// Convert markdown file to HTML
func markdownFileFunc(path string) template.HTML {
	file, err := os.Open("./" + path)
	if err != nil {
		os.Stderr.WriteString("Error to access markdown file: " + path)
		return template.HTML("")
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		os.Stderr.WriteString("Error to read the markdown file")
		return template.HTML("")
	}

	md := markdown.New(markdown.XHTMLOutput(true))
	html := md.RenderToString(bytes)
	return template.HTML(html)
}

func buildAll(source string, destination string, ignoreSpecialFiles bool) error {
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
			err = buildAll(source+"/"+f.Name(), destination+"/"+f.Name(), ignoreSpecialFiles)
			if err != nil {
				return nil
			}
		}
	} else {
		ext := filepath.Ext(source)
		if ext == ".html" {
			if !strings.HasPrefix(source, "template/layouts/") {
				dest, err := os.Create(destination)
				if err != nil {
					return fmt.Errorf("Error to create: %s", destination)
				}
				defer dest.Close()

				err = build(source, dest, nil)
				if err != nil {
					return err
				}
			}
		}

		if ignoreSpecialFiles || !regexp.MustCompile("(?i)(.json|.md|.html)").Match([]byte(source)) {
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
	}

	return nil
}

func build(html string, w io.Writer, httpWriter http.ResponseWriter) error {
	// TODO: How about if the file named *.JSON?
	path := html[:len(html)-5]
	meta := path + ".json"
	_, err := os.Stat(html)
	if os.IsNotExist(err) {
		// Start the build procedure
		return fmt.Errorf("No JSON (meta-data) file found for: %s", html)
	}

	// Load the json file using default params
	data := page{
		Templates:   []string{"layouts/default.html"},
		Preloads:    []string{},
		Language:    "",
		Description: "",
		Keywords:    []string{},
		Author:      "",
		Image:       "",
		Title:       "",
		URL:         "/" + path,
	}

	metaFile, err := ioutil.ReadFile(meta)
	if err != nil {
		return fmt.Errorf("Error to read the JSON meta data: %s", meta)
	}
	err = json.Unmarshal([]byte(metaFile), &data)
	if err != nil {
		return fmt.Errorf("Error to parse the JSON meta data file: %s", meta)
	}

	// Each file can generate from a couple of template files. The default file is
	// 'template/PATH.html' and it's always included by default
	for i := range data.Templates {
		data.Templates[i] = "template/" + data.Templates[i]
	}
	data.Templates = append(data.Templates, path+".html")

	// We provide some extra functionality in our templates files to help generate HTML files easier (ex using Markdown)
	funcMap := template.FuncMap{"StringsJoin": strings.Join, "Markdown": markdownFunc, "MarkdownFile": func(arg string) template.HTML {
		dir := filepath.Dir(path)
		return markdownFileFunc(dir + "/" + arg)
	}}

	// We also suppor preload headers, using a reverse proxy like NGINX you can easily support HTTP2 to boost page loading speed
	if httpWriter != nil {
		for _, preload := range data.Preloads {
			_as := preloads[filepath.Ext(preload)]
			if len(_as) > 0 {
				httpWriter.Header().Add("Link", "<"+preload+">; as="+_as+"; rel=preload")
			}
		}
	}

	tmpl := template.Must(template.New("base").Funcs(funcMap).ParseFiles(data.Templates...))
	err = tmpl.ExecuteTemplate(w, "layout", data)
	if err != nil {
		return fmt.Errorf("Error to build the template: %s", html)
	}
	fmt.Printf("Template compiled: %s\n", html)

	return nil
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
				err := buildAll("template", "build", false)
				if err != nil {
					panic(err)
				}
				err = buildAll("static", "build", true)
				if err != nil {
					panic(err)
				}
			}
		}
		return
	}

	r := mux.NewRouter()
	fs := http.FileServer(http.Dir("static"))

	files, err := ioutil.ReadDir("./static")
	if err != nil {
		os.Stderr.WriteString("Error to read 'static' folder to host 'js' and 'css' files")
	} else {
		for _, f := range files {
			r.PathPrefix("/" + f.Name() + "/").Handler(http.StripPrefix("/", fs))
		}
	}

	// ============================= Handle your rest API here ==============================
	// If you want to have some rest APIs and turn your static website into a fully dynamic
	// Website, you can easily add your handlers here and consume that rest API in your pages
	// ======================================================================================

	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		path := req.URL.Path
		path = "template" + path

		// default page point to 'index' file.
		if strings.HasSuffix(path, "/") {
			path = path + "index"
		}

		// If we don't find the 'json' file, it means there's no page and we just show 404
		file, err := ioutil.ReadFile(path + ".json")
		if err != nil {
			fmt.Println("Resource not found: " + path)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Load the json file using default params
		data := page{
			Templates:   []string{"layouts/default.html"},
			Preloads:    []string{},
			Language:    "",
			Description: "",
			Keywords:    []string{},
			Author:      "",
			Image:       "",
			Title:       "",
			URL:         req.URL.RequestURI(),
		}
		_ = json.Unmarshal([]byte(file), &data)

		// Each file can generate from a couple of template files. The default file is
		// 'template/PATH.html' and it's always included by default
		for i := range data.Templates {
			data.Templates[i] = "template/" + data.Templates[i]
		}
		data.Templates = append(data.Templates, path+".html")

		// We provide some extra functionality in our templates files to help generate HTML files easier (ex using Markdown)
		funcMap := template.FuncMap{"StringsJoin": strings.Join, "Markdown": markdownFunc, "MarkdownFile": func(arg string) template.HTML {
			dir := filepath.Dir(path)
			return markdownFileFunc(dir + "/" + arg)
		}}

		// We also suppor preload headers, using a reverse proxy like NGINX you can easily support HTTP2 to boost page loading speed
		for _, preload := range data.Preloads {
			_as := preloads[filepath.Ext(preload)]
			if len(_as) > 0 {
				w.Header().Add("Link", "<"+preload+">; as="+_as+"; rel=preload")
			}
		}

		tmpl := template.Must(template.New("base").Funcs(funcMap).ParseFiles(data.Templates...))
		tmpl.ExecuteTemplate(w, "layout", data)
	})

	http.Handle("/", r)
	http.ListenAndServe(":3000", nil)
}
