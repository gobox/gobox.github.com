package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Applet struct {
	Name string
	URL  *url.URL
}

const (
	APPLET_PREFIX = "applet-"
)

func AllApplets(c *http.Client) []Applet {
	// FIXME Handle pagination
	r, e := c.Get("https://api.github.com/orgs/gobox/repos")
	if e != nil {
		log.Fatalf("Could not acquire repository list: %s", e)
	}
	defer r.Body.Close()

	var rawrepos []map[string]interface{}
	d := json.NewDecoder(r.Body)
	e = d.Decode(&rawrepos)
	if e != nil {
		log.Fatalf("Could not parse repository list: %s", e)
	}

	applets := make([]Applet, 0, len(rawrepos))
	for _, rawrepo := range rawrepos {
		appletname := rawrepo["name"].(string)
		if !strings.HasPrefix(appletname, APPLET_PREFIX) {
			log.Printf("Skipping %s", appletname)
			continue
		}
		appletname = appletname[len(APPLET_PREFIX):]
		appleturl, e := url.Parse(rawrepo["html_url"].(string))
		if e != nil {
			log.Printf("Repository %s has invalid URL: %s", appletname, e)
			continue
		}
		applets = append(applets, Applet{
			Name: appletname,
			URL:  appleturl,
		})
	}
	return applets
}

func generateAppletIndexFiles(applets []Applet) {
	indextpl := template.Must(template.ParseFiles("tools/importpathgen/applet-index.html.template"))
	e := os.RemoveAll("applet")
	if e != nil {
		log.Fatalf("Could not remove \"applet\": %s", e)
	}
	for _, applet := range applets {
		func() {
			path := filepath.Join("applet", applet.Name)
			e := os.MkdirAll(path, os.FileMode(0755))
			if e != nil {
				log.Printf("Could not create \"%s\": %s", path, e)
				return
			}

			f, e := os.Create(filepath.Join(path, "index.html"))
			if e != nil {
				log.Printf("Could not create index.html in \"%s\": %s", path, e)
				return
			}
			defer f.Close()

			indextpl.Execute(f, applet)
		}()
	}
}

func generateMainIndexFiles(applets []Applet) {
	indextpl := template.Must(template.ParseFiles("tools/importpathgen/main-index.html.template"))
	f, e := os.Create("index.html")
	if e != nil {
		log.Printf("Could not create index.html in \"index.html\": %s", e)
		return
	}
	defer f.Close()

	indextpl.Execute(f, applets)
}

func main() {
	applets := AllApplets(http.DefaultClient)
	generateAppletIndexFiles(applets)
	generateMainIndexFiles(applets)
}
