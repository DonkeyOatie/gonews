package main

import (
	"log"
	"sync"
	"io/ioutil"
	"os"
)

// nuviNewsURL is the URL which stores the zip files containing news articles
const nuviNewsURL = "http://feed.omgili.com/5Rh5AMTrc4Pv/mainstream/posts/"

// create wait group so we can block when routines are running
var wg sync.WaitGroup

func main() {
	// create a tmp dir to store the zip files in
	dir, err := ioutil.TempDir(".", "zips")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// get the complete list of file names available
	fileNames := getFileNames()

	// create a channel for the downloader and processer to communicate
	// with
	c := make(chan string)

	go processFiles(c)
	getFiles(fileNames, c)

}
