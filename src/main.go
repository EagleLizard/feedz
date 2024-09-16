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
	/*
		Map the full xmlns string to its prefix for convenience.
			Required because of how std encoding/xml returns tokens
			when using Decoder.
		See: https://pkg.go.dev/encoding/xml#Name
	*/
	nsMap := make(map[string]string)

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
			fmt.Println("StartElement")
			fmt.Printf("%+v\n", t.Name)
			if len(t.Name.Space) > 0 {
				nsPrefix := nsMap[t.Name.Space]
				fmt.Println(t.Name.Space)
				fmt.Println(nsPrefix)
			}
			for _, attr := range t.Attr {
				// fmt.Printf("Name: %+v\n", attr.Name)
				// fmt.Printf("Value: %v\n", attr.Value)
				if attr.Name.Space == "xmlns" {
					/*
						TODO: remove namespaces from map when an element's end
							element is encountered that defined the space (when it
							falls out of scope)
					*/
					nsMap[attr.Value] = attr.Name.Local
					// fmt.Printf("xmlns:%s=\"%s\"\n", attr.Name.Local, attr.Value)
				}
			}
			// return
		case xml.EndElement:
			// fmt.Println("EndElement")
			// fmt.Printf("%+v\n", t.Name)
		case xml.CharData:
			// fmt.Println("CharData")
		case xml.Comment:
			// fmt.Println("Comment")
		case xml.ProcInst:
			// fmt.Println("ProcInst")
			// fmt.Printf("target: %s\n", t.Target)
			// fmt.Printf("inst: %s\n", t.Inst)
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
