package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/EagleLizard/feedz/src/constants"
	"github.com/EagleLizard/sysmon-go/src/lib/argv"
)

func main() {
	parsedArv := argv.ParseArgv(os.Args)
	fmt.Printf("%v\n", parsedArv)
	fmt.Println("hi ~")
	switch parsedArv.Cmd {
	case "fetch":
		fetchFeeds()
	case "decode":
		fallthrough
	case "d":
		decode()
	default:
		fmt.Printf("Cmd: %s\n", parsedArv.Cmd)
	}
}

func decode() {
	feedFilePath := constants.TestFeedFilePath
	fmt.Printf("Decoding: %s\n", feedFilePath)
	r, err := os.Open(feedFilePath)
	if err != nil {
		log.Fatal(err)
	}
	d := xml.NewDecoder(r)
	t, tokenErr := d.Token()
	if tokenErr != nil {
		log.Fatal(tokenErr)
	}
	fmt.Printf("%+v\n", t)
}

func fetchFeeds() {
	feedUris := getFeedUris()
	for _, feedUri := range feedUris {
		fmt.Println(feedUri)
		resp, err := http.Get(feedUri)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(body))
	}
}

func getFeedUris() []string {
	f, err := os.Open(constants.TestFeedsFilePath)
	if err != nil {
		log.Fatal(err)
	}
	sc := bufio.NewScanner(f)
	feedUris := []string{}
	for sc.Scan() {
		line := sc.Text()
		if len(line) > 0 {
			feedUris = append(feedUris, strings.TrimSpace(line))
		}
	}
	return feedUris
}
