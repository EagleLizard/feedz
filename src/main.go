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
		decodeCmd()
	default:
		fmt.Printf("Cmd: %s\n", parsedArv.Cmd)
	}
}

type rssParseStateEnum int

const (
	initRss rssParseStateEnum = iota
	rssTag
	rssChannel
)

func decodeCmd() {
	feedFilePath := constants.TestFeedFilePath
	fmt.Printf("Decoding: %s\n", feedFilePath)
	r, err := os.Open(feedFilePath)
	if err != nil {
		log.Fatal(err)
	}
	d := xml.NewDecoder(r)
	/*
		Map the full xmlns string to its prefix for convenience.
			Required because of how std encoding/xml returns tokens
			when using Decoder.
		See: https://pkg.go.dev/encoding/xml#Name
	*/
	nsMap := make(map[string]string)
	tagStack := []xml.StartElement{}
	rssParseState := initRss
	for {
		t, tokenErr := d.Token()
		if tokenErr != nil {
			if tokenErr == io.EOF {
				break
			}
			log.Fatal(tokenErr)
		}
		switch t := t.(type) {
		case xml.StartElement:
			for _, attr := range t.Attr {
				if attr.Name.Space == "xmlns" {
					/*
						TODO: remove namespaces from map when an element's end
							element is encountered that defined the space (when it
							falls out of scope)
					*/
					nsMap[attr.Value] = attr.Name.Local
				} else {
					switch rssParseState {
					case rssTag:
						fmt.Printf("%+v\n", attr)
						if attr.Name.Local == "version" {
							fmt.Printf("version: %s\n", attr.Value)
						} else {
							log.Fatal(fmt.Errorf("invalid rss attribute: %+v", attr))
						}
					}
				}
			}
			switch rssParseState {
			case initRss:
				fmt.Printf("%+v\n", t)
				return
			}
			tagStack = append(tagStack, t)
			// return
		case xml.EndElement:
			if len(tagStack) > 0 {
				topTag := tagStack[len(tagStack)-1]
				if topTag.Name.Space == t.Name.Space &&
					topTag.Name.Local == t.Name.Local {
					tagStack = tagStack[:len(tagStack)-1]
					if nsMap[topTag.Name.Space] != "" {
						fmt.Printf("%+v\n", nsMap[topTag.Name.Space])
						// fmt.Printf("%+v\n", topTag)
					}
				}
			}
			fmt.Println("EndElement")
			// fmt.Printf("%+v\n", t.Name)
		case xml.CharData:
			// fmt.Println("CharData")
			// fmt.Printf("%s\n", t)
			switch rssParseState {
			case initRss:
				/*
					should be whitespace, do nothing
				*/
			}
		case xml.Comment:
			// fmt.Println("Comment")
		case xml.ProcInst:
			// fmt.Println("ProcInst")
			// fmt.Printf("%+v\n", t)
			// fmt.Printf("Inst: %s", t.Inst)
			// fmt.Printf("target: %s\n", t.Target)
			/*
				should be something like:
					version="1.0" encoding="UTF-8"
			*/
			switch rssParseState {
			case initRss:
				rssParseState = rssTag
			default:
				log.Fatalf("invalid ProcInst: %+v", t)
			}
		case xml.Directive:
			// fmt.Println("Directive")
		}
	}
	// fmt.Printf("%+v\n", t)
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
