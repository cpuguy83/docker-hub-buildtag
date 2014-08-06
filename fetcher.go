package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/hoisie/redis"
)

var redisUrl = flag.String("redis", "localhost:6379", "Redis URL <host ip>:<port>")
var serveAddr = flag.String("serve", ":80", "Interface address and port, <ip>:<port>")
var keyTimeout = flag.Int("cache-timeout", 300, "Time, in seconds, for key expiration")
var assetsDir = flag.String("assets", "/fetcher", "Dir which has assets stored in it")

var Redis redis.Client

func main() {
	flag.Parse()
	Redis = redis.Client{Addr: *redisUrl}
	http.HandleFunc("/", handler)
	http.ListenAndServe(*serveAddr, Log(http.DefaultServeMux))
}

func Log(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func handler(w http.ResponseWriter, r *http.Request) {
	repo := r.URL.Path[1:]

	status, err := cache(repo)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s", err), 500)
	} else {
		status = *assetsDir + "/" + status + ".svg"
		w.Header().Set("Content-Type", "image/svg+xml")
		http.ServeFile(w, r, status)
	}

}

func GetBuildStatus(repo string) (string, error) {
	hub_url := "https://registry.hub.docker.com"
	url := fmt.Sprintf("%s/u/%s", hub_url, repo)

	build_uri, err := getBuildUri(url)
	if err != nil {
		return "", err
	}

	url = fmt.Sprintf("%s/%s", hub_url, build_uri)
	return getBuildStatus(url)
}

func cache(key string) (string, error) {
	rkey := "hub_repo_status:" + key
	val, err := Redis.Get(rkey)
	if err != nil {
		s, err := GetBuildStatus(key)
		if err != nil {
			return "", err
		}

		val = []byte(s)
		go func() {
			Redis.Set(rkey, val)
			Redis.Expire(rkey, int64(*keyTimeout))
		}()
		return string(val), nil
	}
	return string(val), nil
}

func getBuildUri(src string) (string, error) {
	var (
		href string
		err  error
	)
	doc, err := goquery.NewDocument(src)
	if err != nil {
		return "", err
	}
	doc.Find("a:contains(\"Build Details\")").Each(func(i int, s *goquery.Selection) {
		var exists bool
		href, exists = s.Attr("href")
		if !exists {
			err = fmt.Errorf("Could not get build details")
		}
	})

	return href, err
}

func getBuildStatus(src string) (string, error) {
	doc, err := goquery.NewDocument(src)
	if err != nil {
		return "", err
	}
	cssPath := "#repo-info-tab > div.repository > table > tbody tr > td"
	var status string
	doc.Find(cssPath).Each(func(i int, s *goquery.Selection) {
		txt := s.Text()
		if txt == "Finished" || txt == "Error" {
			if status == "" {
				switch txt {
				case "Finished":
					status = "passing"
				case "Error":
					status = "failing"
				default:
					status = txt
				}

			}
		}
	})

	return status, nil
}
