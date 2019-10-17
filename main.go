package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/anaskhan96/soup"
	aw "github.com/deanishe/awgo"
)

func getURLs(url string) []string {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var urls []string
	err = json.Unmarshal(body, &urls)
	if err != nil {
		log.Fatal(err)
	}
	return urls
}

func validateURL(urls []string) string {
	for _, url := range urls {
		resp, err := http.Get(url)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode == 200 && resp.StatusCode <= 299 {
			return url
		}
		panic(fmt.Sprintf("There is no validated url"))
	}

	return ""
}

func getLink(url string) string {
	resp, err := soup.Get(url)
	if err != nil {
		os.Exit(1)
	}
	doc := soup.HTMLParse(resp)
	linkRoot := doc.Find("div", "id", "article")
	if linkRoot.Error != nil {
		return ""
	}
	link := linkRoot.Find("iframe")
	if len(link.Attrs()) != 0 {
		mirror := strings.Split(link.Attrs()["src"], "#")[0]
		if strings.HasPrefix(mirror, "//") {
			mirror = "http:" + mirror
		}
		return mirror
	}
	return ""
}

func downloadFile(url string) {
	resp, err := http.Get(url)
	if err != nil {
		os.Exit(1)
	}
	defer resp.Body.Close()
	if resp.Header.Get("content-type") == "application/pdf" {
		temp := strings.Split(url, "/")
		PdfName := temp[len(temp)-1]
		filepath := os.Getenv("HOME") + "/Downloads/" + PdfName
		out, err := os.Create(filepath)
		if err != nil {
			panic(fmt.Sprintf("Failed to create file"))
		}
		defer out.Close()

		_, err = io.Copy(out, resp.Body)
		if err != nil {
			panic(fmt.Sprintf("Failed to write file"))
		}
	} else {
		err := exec.Command("open", url).Start()
		if err != nil {
			panic(fmt.Sprintf("Failed to open url in browser"))
		}
	}
}

var wf *aw.Workflow

func init() {
	wf = aw.New()
}

func run() {
	wf.Args() // call to handle any magic actions
	flag.Parse()
	query := ""
	if args := flag.Args(); len(args) > 0 {
		query = args[0]
	}
	if query == "" {
		panic(fmt.Sprintf("query is empty"))
	}
	log.Printf("[main] query=%s", query)
	SciHubURL := validateURL(getURLs("https://whereisscihub.now.sh/api"))
	paperURL := SciHubURL + "/" + query
	link := getLink(paperURL)
	if link == "" {
		err := exec.Command("open", paperURL).Start()
		if err != nil {
			panic(fmt.Sprintf("Failed to open url in browser"))
		}
		return
	}
	downloadFile(link)
	wf.NewItem("Download Successfully")
	wf.SendFeedback()
}

func main() {
	wf.Run(run)
}
