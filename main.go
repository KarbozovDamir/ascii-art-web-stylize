package main

import (
	"bufio"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"text/template"
)

var Tmp *template.Template
var err error

func init() {
	Tmp, err = template.ParseGlob("templates/*.html")
}

type Page struct {
	Input  string
	Output string
	Fs     string
	Btn    string
}

func Valid(s string) bool {
	for _, v := range s {
		if v == '\r' || v == '\n' {
			continue
		}
		if v < 32 || 126 < v {
			return false
		}
	}
	return true
}

func GetMethod(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}
	if r.URL.Path != "/" {
		http.Error(w, http.StatusText(404), http.StatusNotFound)
		return
	}
	Tmp.ExecuteTemplate(w, "Index.html", nil)
}

func PostMethod(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}
	values, err := url.ParseQuery(string(bytes))
	if err != nil {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}
	p := Page{}
	if _, ok := values["input"]; !ok {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	} else if _, ok2 := values["fs"]; !ok2 {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}
	for i, v := range values {
		switch i {
		case "input":
			p.Input = v[0]
		case "fs":
			p.Fs = v[0]
		case "button":
			p.Btn = v[0]
		}
	}

	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}
	if r.URL.Path != "/ascii-art/" {
		http.Error(w, http.StatusText(404), http.StatusNotFound)
		return
	}
	if (p.Fs != "standard" && p.Fs != "shadow" && p.Fs != "thinkertoy") || !Valid(p.Input) {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}
	p.Output, err = GetArt(p.Input, p.Fs)
	// button := r.FormValue("button")
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	switch p.Btn {
	case "submit":
		Tmp.ExecuteTemplate(w, "Index.html", p)
	case "export":
		file := strings.NewReader(p.Output)
		// filesize := strconv.FormatInt(file.Size(), 10)
		filesize := strconv.Itoa(int(file.Size()))
		w.Header().Set("Content-Type", "text")
		w.Header().Set("Content-Disposition", "attachment; filename=TheFile.txt")
		w.Header().Set("Content-Length", filesize)
		file.Seek(0, 0)
		io.Copy(w, file)
	}
}

func main() {
	log.Printf("Server: http://%v", "localhost:8080")

	http.Handle("/templates/", http.StripPrefix("/templates/", http.FileServer(http.Dir("templates"))))

	http.HandleFunc("/", GetMethod)
	http.HandleFunc("/ascii-art/", PostMethod)

	log.Println(http.ListenAndServe(":8080", nil))
}

func GetArt(s, fs string) (string, error) {
	words := strings.Split(s, "\r\n")
	file, err := os.Open("Fonts/" + fs + ".txt")
	if err != nil {
		return "", err
	}
	content := bufio.NewScanner(file)
	arr := []string{}
	for content.Scan() {
		arr = append(arr, content.Text())
	}
	t := make([][8]string, len(words))
	for i := 0; i < len(words); i++ {
		for k := 0; k < 8; k++ {
			for j := 0; j < len(words[i]); j++ {
				num := (int(words[i][j])-32)*9 + 1
				t[i][k] += arr[num+k]
			}
		}
	}
	ans := ""
	for i := 0; i < len(words); i++ {
		for k := 0; k < 8; k++ {
			ans += t[i][k] + "\n"
		}
	}
	return ans, nil
}
