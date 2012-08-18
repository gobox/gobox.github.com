package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
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

func main() {
	log.Printf("Repos: %#v", AllApplets(http.DefaultClient))
}
