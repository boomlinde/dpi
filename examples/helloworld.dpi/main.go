package main

import (
	"fmt"
	"github.com/boomlinde/dpi"
	"io"
	"log"
	"os"
)

func main() {
	log.Println("starting", os.Args[0])
	log.Fatal(dpi.AutoRun(func(tag map[string]string, w io.Writer) error {
		log.Println("got tag", tag)
		switch tag["cmd"] {
		case "open_url":
			log.Println("opening url")
			dpi.Tag(w, map[string]string{
				"cmd": "start_send_page",
				"url": tag["url"],
			})
			w.Write([]byte("Content-Type: text/html\r\n\r\n"))
			fmt.Fprintf(w, "<h1>Hello world %s</h1>\n", tag["url"])
			return dpi.Done
		case "DpiBye":
			log.Println("bye!")
			os.Exit(0)
		}
		return nil
	}))
}
