package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"

	"golang.org/x/net/html"
)

type config struct {
	username     string
	userpassword string
	url          string
}

func main() {
	var cfg config
	otmProcesses := [][]string{}
	otmProcess := make([]string, 11)

	flag.StringVar(&cfg.username, "u", "DBA.ADMIN", "User name")
	flag.StringVar(&cfg.userpassword, "p", "", "User password")
	flag.StringVar(&cfg.url, "url", "", "URL")
	flag.Parse()

	if cfg.url == "" {
		fmt.Println("URL is required")
		usage()
	}

	if cfg.userpassword == "" {
		cfg.userpassword = os.Getenv("OTMPWD")
		if cfg.userpassword == "" {
			fmt.Println("User password is required")
			usage()
		}
	}

	CookieJar, _ := cookiejar.New(nil)
	otmClient := &http.Client{
		CheckRedirect: nil,
		Jar:           CookieJar,
	}

	fmt.Println("Connecting to ", cfg.url, "...")
	response, _ := otmClient.PostForm(cfg.url+"/GC3/glog.webserver.servlet.umt.Login", url.Values{"username": {cfg.username}, "userpassword": {cfg.userpassword}})
	fmt.Println("Getting inromation about open processes...")
	response, _ = otmClient.Get(cfg.url + "/GC3/glog.webserver.process.walker.ProcessWalkerDiagServlet")

	tokens := html.NewTokenizer(response.Body)

	for {
		tokenType := tokens.Next()
		if tokenType == html.ErrorToken {
			break
		}
		if tokenType == html.StartTagToken {
			// Looking for "td" tags with specific attributes
			tagName, _ := tokens.TagName()
			if string(tagName) == "tr" {
				_, tagAttr, _ := tokens.TagAttr()
				if string(tagAttr) == "gridColGroupRow" {
					prcCounter := 0
					for {
						innerTokenType := tokens.Next()
						innerToken := tokens.Token()
						if innerTokenType == html.EndTagToken && innerToken.Data == "tr" {
							break
						}
						if innerTokenType == html.StartTagToken && innerToken.Data == "a" {
							tokens.Next()
						}
						if innerTokenType == html.TextToken {
							otmProcess[prcCounter] = innerToken.Data
							prcCounter++
						}
					}
					otmProcesses = append(otmProcesses, otmProcess)
				}
			}
		}
	}

	fmt.Println("+-------------------------------------------------+--------------+-------------+")
	fmt.Println("| Description                                     | Server       | Waited      |")
	fmt.Println("+-------------------------------------------------+--------------+-------------+")
	for i := range otmProcesses {
		fmt.Printf("| %-48s| %-13s| %-12s|\n", otmProcesses[i][0], otmProcesses[i][3], otmProcesses[i][5])
	}
	fmt.Println("+-------------------------------------------------+--------------+-------------+")
	fmt.Println("Total: ", len(otmProcesses))
}

func usage() {
	fmt.Printf("Usage: otmproc [-u <username>] [-p <password>] -url <OTM URL>\n")
	fmt.Printf("  -u <username>  - OTM user name, defaults to DBA.ADMIN\n")
	fmt.Printf("  -p <password>  - OTM user password\n")
	fmt.Printf("                   Password can be passed via environment variable OTMPWD\n")
	fmt.Printf("  -url <OTM URL> - OTM URL address\n")
	os.Exit(0)
}
