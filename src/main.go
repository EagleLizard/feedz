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
	rssChannelTag
	rssChannelInner
	rssChannelItem
	rssItemTitle
	rssItemDesc
)

/*
	An item must contain a title OR a description
		all other elements are optional
	Child elements:
		title: cdata
		description: cdata
	see: https://www.rssboard.org/rss-profile#element-channel-item
*/

type RssItem struct {
	title       string
	description string
}

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

	rssItems := []RssItem{}
	var currRssItem *RssItem

	getTagStr := func(el xml.StartElement) string {
		nsStr := nsMap[el.Name.Space]
		var tagPrefix string
		if len(nsStr) > 0 {
			tagPrefix = nsStr + ":"
		}
		return tagPrefix + el.Name.Local
	}

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
				}
			}
			switch rssParseState {
			case rssTag:
				if t.Name.Local != "rss" {
					log.Fatalf("invalid rss tag: %+v", t)
				}
				rssParseState = rssChannelTag
			case rssChannelTag:
				if t.Name.Local != "channel" {
					log.Fatalf("%v: invalid StartElement (expected channel): %+v", rssParseState, t)
				}
				rssParseState = rssChannelInner
			case rssChannelInner:
				if t.Name.Local == "item" {
					if currRssItem != nil {
						panic(fmt.Sprintf("unexpected item open tag. currRssItem: %+v, currTag: %+v", currRssItem, t))
						// log.Fatalf("unexpected item open tag. currRssItem: %+v, currTag: %+v", currRssItem, t)
					}
					currRssItem = &RssItem{}
					rssParseState = rssChannelItem
				}
			case rssChannelItem:
				if len(nsMap[t.Name.Space]) > 0 {
					/*
						TODO: do something with namespace elements
					*/
				} else {
					switch t.Name.Local {
					case "title":
						rssParseState = rssItemTitle
					case "description":
						rssParseState = rssItemDesc
					}
				}
			}
			tagStack = append(tagStack, t)

		case xml.EndElement:
			var startEl xml.StartElement
			if len(tagStack) > 0 {
				startEl = tagStack[len(tagStack)-1]
				if startEl.Name.Space == t.Name.Space &&
					startEl.Name.Local == t.Name.Local {
					tagStack = tagStack[:len(tagStack)-1]
				}
			} else {
				panic("empty tagStack (should be unreachable code)")
			}

			var topEl *xml.StartElement
			if len(tagStack) > 0 {
				topEl = &tagStack[len(tagStack)-1]
			}
			switch rssParseState {
			case rssChannelInner:
				tagStr := getTagStr(startEl)
				fmt.Print("\n")
				fmt.Printf("%s\n", tagStr)
				fmt.Printf("%+v\n", topEl)
				if t.Name.Local == "item" {
					return
				}
			case rssChannelItem:
				switch startEl.Name.Local {
				case "item":
					if currRssItem == nil {
						panic(fmt.Sprintf("unexpected nil currRssItem, end tag: %+v\n", t))
					}
					rssItems = append(rssItems, *currRssItem)
					currRssItem = nil
					rssParseState = rssChannelInner
				}
			case rssItemTitle:
				fallthrough
			case rssItemDesc:
				rssParseState = rssChannelItem
			default:
				log.Fatalf("%v: invalid EndElement: %+v", rssParseState, t)
			}
		case xml.CharData:
			switch rssParseState {
			case initRss:
				/*
					should be whitespace, do nothing
				*/
			case rssItemTitle:
				currCDataStr := string(t)
				if len(strings.TrimSpace(currCDataStr)) > 0 {
					currRssItem.title += currCDataStr
				}
			case rssItemDesc:
				currCDataStr := string(t)
				if len(strings.TrimSpace(currCDataStr)) > 0 {
					currRssItem.description += currCDataStr
				}
			}
		case xml.Comment:
			fmt.Println("Comment")
		case xml.ProcInst:
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

	fmt.Printf("len(rssItems): %d\n", len(rssItems))
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
