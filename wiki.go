// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

//Page holds Title and Body as bytes
type Page struct {
	Title  string
	Body   []byte
	Image1 string
	Image2 string
	Image3 string
	Image4 string
	Image5 string
	Image6 string
	Image7 string
	Image8 string
	Image9 string
}

func backup(title string) error {
	p, err := loadPage(title)
	if err != nil {
		return err
	}
	filename := p.Title + time.Now().Local().Format("2006_01_02_15_04_05") + ".bak"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body,
		Image1: "/images/" + title + "Image1.jpg?" + time.Now().Local().Format("2006_01_02_15_04_05"),
		Image2: "/images/" + title + "Image2.jpg?" + time.Now().Local().Format("2006_01_02_15_04_05"),
		Image3: "/images/" + title + "Image3.jpg?" + time.Now().Local().Format("2006_01_02_15_04_05"),
		Image4: "/images/" + title + "Image4.jpg?" + time.Now().Local().Format("2006_01_02_15_04_05"),
		Image5: "/images/" + title + "Image5.jpg?" + time.Now().Local().Format("2006_01_02_15_04_05"),
		Image6: "/images/" + title + "Image6.jpg?" + time.Now().Local().Format("2006_01_02_15_04_05"),
		Image7: "/images/" + title + "Image7.jpg?" + time.Now().Local().Format("2006_01_02_15_04_05"),
		Image8: "/images/" + title + "Image8.jpg?" + time.Now().Local().Format("2006_01_02_15_04_05"),
		Image9: "/images/" + title + "Image9.jpg?" + time.Now().Local().Format("2006_01_02_15_04_05")}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		//w.Write([]byte(title + " " + err.Error()))
		return
	}
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func tocHandler(w http.ResponseWriter, r *http.Request, title string) {
	var body = `
	<head>
	<script type="text/javascript">
		function startnew(s, e){
			var newname = document.getElementById("newname").value;
			if (newname == ""){
				alert("Name must not be blank"); return;
			}
			var re = /^[a-zA-Z0-9]{1,200}$/
			if (!re.test(newname)){
				alert("You must have only letters and digits in the file name, and length must be less than 201"); return;
			}
			window.location.href = "/view/" + newname;
		}
	</script>
	</head>
	<h1>Bob's Quick and Simple Go Wiki Yourself</h1>
		<h3>Table of Contents<h3>
		</br></br>
		<button type="button" id="doit" onclick="startnew()">Create New Page</button>
		<input type="text" id="newname" value="" 
			onkeypress="javascript: if (event.keyCode == 13) document.getElementById('doit').click();" >
		</br></br>
		`
	files, _ := ioutil.ReadDir("./")
	for _, f := range files {
		var fname = f.Name()
		if strings.HasSuffix(fname, ".txt") {

			body += "<p>[<a href=\"/view/" + strings.TrimSuffix(fname, ".txt") + "\">" + strings.TrimSuffix(fname, ".txt") + "</a>]</p>"
		}
	}
	w.Write([]byte(body))
	//renderTemplate(w, "toc", &Page{Title: "", Body: []byte(body)})
}

func deleteHandler(w http.ResponseWriter, r *http.Request, title string) {
	err := backup(title)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = os.Remove(title + ".txt")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
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

func uploadHandler(w http.ResponseWriter, r *http.Request, title string, imageNumber string) {
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		http.Error(w, "ParseMultipartForm: "+err.Error(), http.StatusInternalServerError)
		return
	}
	file, _, err := r.FormFile("uploadfile" + imageNumber)
	if err != nil {
		http.Error(w, "FormFile for uploadfile"+imageNumber+": "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()
	//fmt.Fprintf(w, "%v", handler.Header)
	f, err := os.OpenFile("./images/"+title+"Image"+imageNumber+".jpg", os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		http.Error(w, "OpenFile: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()
	io.Copy(f, file)
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

var templates = template.Must(template.ParseFiles("edit.html", "view.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var validPath = regexp.MustCompile("/((edit|save|view|toc|delete|upload)/(([a-zA-Z0-9]+)(/([0-9]))*)*)*$")

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[3])
	}
}
func makeUploadHandler(fn func(http.ResponseWriter, *http.Request, string, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[4], m[6])
	}
}

func main() {
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("images"))))
	http.HandleFunc("/upload/", makeUploadHandler(uploadHandler))
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.HandleFunc("/toc/", makeHandler(tocHandler))
	http.HandleFunc("/delete/", makeHandler(deleteHandler))
	http.HandleFunc("/", makeHandler(tocHandler))

	http.ListenAndServe(":8099", nil)
}
