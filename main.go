package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/anaskhan96/soup"
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

func fileExists(filepath string) bool {
	if _, err := os.Stat(filepath); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func downloadFile(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		os.Exit(1)
	}

	defer resp.Body.Close()

	if resp.Header.Get("content-type") == "application/pdf" {
		temp := strings.Split(url, "/")
		PdfName := temp[len(temp)-1]
		filepath := os.Getenv("HOME") + "/Downloads/" + PdfName
		if fileExists(filepath) {
			return "file exists"
		}
		out, err := os.Create(filepath)
		if err != nil {
			panic(fmt.Sprintf("Failed to create file"))
		}
		defer out.Close()

		_, err = io.Copy(out, resp.Body)
		if err != nil {
			panic(fmt.Sprintf("Failed to write file"))
		}
		return "success"
	}

	err = exec.Command("open", url).Start()
	if err != nil {
		panic(fmt.Sprintf("Failed to download file and open url in browser"))
	}
	return "fail"
}

func main() {
	query := os.Args[1]
	SciHubURL := validateURL(getURLs("https://whereisscihub.now.sh/api"))
	paperURL := SciHubURL + "/" + query
	link := getLink(paperURL)
	if link == "" {
		err := exec.Command("open", paperURL).Start()
		if err != nil {
			panic(fmt.Sprintf("Failed to open url in browser"))
		}
		fmt.Println("Failed to download, try to open URL in your browser")
		return
	}
	message := downloadFile(link)
	if message == "success" {
		fmt.Println("Download finished")
	} else if message == "file exists" {
		fmt.Println("Failed to download, File already exists in Download directory")
	} else {
		fmt.Println("Failed to download, try to open URL in your browser")
	}
}
