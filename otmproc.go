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

var (
	username     string
	userpassword string
	otmUrl       string
)

type otmProcess struct {
	
}

func main() {
	otmProcesses := [][]string{}
	otmProcess := make([]string, 11)

	flag.StringVar(&username, "u", "DBA.ADMIN", "User name")
	flag.StringVar(&userpassword, "p", "", "User password")
	flag.StringVar(&otmUrl, "url", "", "URL")
	flag.Parse()

	if otmUrl == "" {
		fmt.Println("URL is required")
		usage()
	}

	if userpassword == "" {
		userpassword = os.Getenv("OTMPWD")
		if userpassword == "" {
			fmt.Println("User password is required")
			usage()
		}
	}

	CookieJar, _ := cookiejar.New(nil)
	otmClient := &http.Client{
		CheckRedirect: nil,
		Jar:           CookieJar,
	}

	fmt.Println("Connecting to ", otmUrl, "...")
	response, _ := otmClient.PostForm(otmUrl+"/GC3/glog.webserver.servlet.umt.Login", url.Values{"username": {username}, "userpassword": {userpassword}})
	fmt.Println("Getting inromation about open processes...")
	response, _ = otmClient.Get(otmUrl + "/GC3/glog.webserver.process.walker.ProcessWalkerDiagServlet")

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
					fmt.Println(otmProcesses[0])
				}
			}
		}
	}

	fmt.Println("+-------------------------------------------------+--------------+-------------+")
	fmt.Println("| Description                                     | Server       | Waited      |")
	fmt.Println("+-------------------------------------------------+--------------+-------------+")
	for i := range otmProcesses {
		fmt.Println(otmProcesses[i])
		//fmt.Printf("| %-48s| %-13s| %-12s|\n", otmProcesses[i][0], otmProcesses[i][3], otmProcesses[i][5])
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
