package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
)

var endpoints map[string]string

func loadConfig(path string) {
	dat, err := os.ReadFile(path)

	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(dat, &endpoints)

	if err != nil {
		panic(err)
	}
}

func fetch(url string, headers map[string]string) (string, *http.Response, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		panic(err)
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	r, err := client.Do(req)

	if err != nil {
		return "", nil, err
	}

	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		panic(err)
	}

	return string(body), r, nil
}

func readme(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "IPTV Playlist Basic Rewriter\n")
}

func rewrite(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/rewrite/")

	url := ""
	for k, v := range endpoints {
		println(path, k, v)
		if k == path {
			url = v
			break
		}
	}

	if url == "" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not found"))

		return
	}

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("GET ONLY"))

		return
	}

	username, password, ok := r.BasicAuth()

	if !ok {
		w.Header().Set("WWW-Authenticate", `Basic realm="Rewrite URLs with credential"`)
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "%d Not authorized.", http.StatusUnauthorized)

		return
	}

	authorization := ""

	if val, ok := r.Header["Authorization"]; ok {
		authorization = val[0]
	}

	useragent := "IPTVPlaylistBasicRewriter 1.0.0"

	if val, ok := r.Header["User-Agent"]; ok {
		useragent = fmt.Sprintf("%s %s", val[0], useragent)
	}

	body, res, err := fetch(url, map[string]string{"Authorization": authorization, "User-Agent": useragent})

	if err != nil {
		panic(err)
	}

	if val, ok := res.Header["Content-Type"]; ok {
		w.Header().Set("Content-Type", val[0])
	}
	w.WriteHeader(res.StatusCode)

	re := regexp.MustCompile(`(https?://)([-a-zA-Z0-9()@:%_\+.~#?&//=]*)`)
	fmt.Fprintf(w, re.ReplaceAllString(body, fmt.Sprintf("${1}%s:%s@$2", username, password)))
}

func main() {
	loadConfig("./config.json")

	http.HandleFunc("/", readme)
	http.HandleFunc("/rewrite/", rewrite)

	http.ListenAndServe(":8090", nil)
}
