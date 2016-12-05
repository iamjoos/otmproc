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
	otmUser     string
	otmPassword string
	otmURL      string
)

func main() {
	otmProcesses := [][11]string{}
	otmProcess := [11]string{}

	flag.StringVar(&otmUser, "u", "DBA.ADMIN", "User name")
	flag.StringVar(&otmPassword, "p", "", "User password")
	flag.StringVar(&otmURL, "url", "", "URL")
	flag.Parse()

	if otmURL == "" {
		fmt.Println("URL is required")
		usage()
	}

	if otmPassword == "" {
		otmPassword = os.Getenv("OTMPWD")
		if otmPassword == "" {
			fmt.Println("User password is required")
			usage()
		}
	}

	CookieJar, _ := cookiejar.New(nil)
	otmClient := &http.Client{
		CheckRedirect: nil,
		Jar:           CookieJar,
	}

	fmt.Println("Connecting to ", otmURL, "...")
	response, _ := otmClient.PostForm(otmURL+"/GC3/glog.webserver.servlet.umt.Login", url.Values{"username": {otmUser}, "userpassword": {otmPassword}})
	fmt.Println("Getting inromation about open processes...")
	response, _ = otmClient.Get(otmURL + "/GC3/glog.webserver.process.walker.ProcessWalkerDiagServlet")

	tokens := html.NewTokenizer(response.Body)

	for {
		tokenType := tokens.Next()
		if tokenType == html.ErrorToken {
			break
		}
		if tokenType == html.StartTagToken {
			// Looking for "tr" tags...
			tagName, _ := tokens.TagName()
			if string(tagName) == "tr" {
				_, tagAttr, _ := tokens.TagAttr()
				// ...with specific attributes
				if string(tagAttr) == "gridColGroupRow" {
					prcCounter := 0
					for {
						innerTokenType := tokens.Next()
						innerToken := tokens.Token()
						// Exit loop if </tr> reached
						if innerTokenType == html.EndTagToken && innerToken.Data == "tr" {
							break
						}
						// Skip <a> tags
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
