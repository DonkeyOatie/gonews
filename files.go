package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/net/html"
)

// getFileNames gets the file names as a slice of strings from the index.html
// at nuviNewsURL
func getFileNames() []string {
	var files []string

	resp, err := http.Get(nuviNewsURL)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	tk := html.NewTokenizer(resp.Body)

	// This is pretty messy.  In fact it is hideous.  In short, we are
	// iterating through the html elements in the index.html to find the
	// table, and then the links within the table so we can get the files
	// name (the text values in link)
	for tk.Token().Data != "html" {
		// we are past the doctype, lets get into the HTML
		tt := tk.Next()
		if tt == html.StartTagToken {
			t := tk.Token()
			// once we hit the table, step inside it
			if t.Data == "td" {
				inner := tk.Next()
				if inner == html.StartTagToken {
					inner = tk.Next()
					// we are now at the actual values,
					// make sure its a file name and not a
					// link to the parent directory
					if inner == html.TextToken {
						value := (string)(tk.Text())
						if strings.Contains(value, "zip") {
							t := strings.TrimSpace(value)
							// good job, return it for processing
							files = append(files, t)
						}
					}
				}
			}
		}
	}
	return files
}

// getFiles iterates over all file names in the list, and if it is a new file
// starts a routine to download it
func getFiles(dir string, fileNames []string, c chan string) {
	// get a file name from the channel
	for i, file := range fileNames {
		wg.Add(1)
		go getFile(dir, file, c)
		// only start 4 routines for fetching files at any one time
		if i > 0 && i%4 == 0 {
			wg.Wait()
		}
	}
	// finally wait incase we have a number of files that is not a multiple
	// of 4
	wg.Wait()
	close(c)
}

// getFile downloads the file with name fileName and once downloaded, puts the
// name in the channel so processing can begin by whatever is listening to the
// channel
func getFile(dir string, fileName string, c chan string) {
	// get the file if the key is not in redis already

	// download the file
	out, err := os.Create(filepath.Join(dir, fileName))
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	resp, err := http.Get(nuviNewsURL + fileName)
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	// once download has finished, add name to channel
	// so it can be processed
	c <- fileName
}

// processFiles listens to the channel c and when a file name is added, it
// begins processing the file -> unzip, extract, tell redis to store
func processFiles(c chan string) {
	for file := range c {
		// process file
		fmt.Println(file)
		// decrement wait group counter, we have finished with this
		// file
		wg.Done()
	}
}
