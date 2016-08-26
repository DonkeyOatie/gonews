package main

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/net/html"
)

type archive struct {
	UnixEpoch string
	Document  document
}

type document struct {
	Type            string  `xml:"type"`
	Forum           string  `xml:"forum"`
	ForumTitle      string  `xml:"forum_title"`
	DiscussionTitle string  `xml:"discussion_title"`
	Language        string  `xml:"language"`
	GMTOffset       string  `xml:"gmt_offset"`
	TopicURL        string  `xml:"topic_url"`
	TopicText       string  `xml:"topic_text"`
	SpamScore       float64 `xml:"spam_score"`
	PostNum         int     `xml:"post_num"`
	PostID          string  `xml:"post_id"`
	PostDate        string  `xml:"post_date"`
	PostTime        string  `xml:"post_time"`
	Username        string  `xml:"username"`
	Post            string  `xml:"post"`
	Signature       string  `xml:"signature"`
	ExternalLinks   string  `xml:"external_links"`
	Country         string  `xml:"country"`
	MainImage       string  `xml:"main_image"`
}

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
		// get the file if the key is not in redis already
		unixEpoch := strings.Split(file, ".")[0]
		fmt.Println(unixEpoch)
		if !isKeyPresent(unixEpoch) {
			wg.Add(1)
			go getFile(dir, file, c)
			// only start 4 routines for fetching files at any one time
			if i > 0 && (i+1)%4 == 0 {
				wg.Wait()
			}
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
	// download the file
	filePath := filepath.Join(dir, fileName)
	out, err := os.Create(filePath)
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
	c <- filePath
}

// processFiles listens to the channel c and when a file name is added, it
// begins processing the file -> unzip, extract, tell redis to store
func processFiles(c chan string) {
	for filePath := range c {
		// process file -> unzip and extract
		filePathParts := strings.Split(filePath, "/")
		dir := filePathParts[0]
		file := filePathParts[1]
		unixEpoch := strings.Split(file, ".")[0]
		outDir := filepath.Join(dir, unixEpoch)

		unzipFile(filePath, outDir)

		// iterate over all files in the outDir and save each one to
		// redis
		fileList := getXMLFileNames(outDir)
		for _, file := range fileList {
			xmlFile, err := os.Open(file)
			if err != nil {
				log.Fatal(err)
			}

			b, _ := ioutil.ReadAll(xmlFile)

			var d document
			xml.Unmarshal(b, &d)

			var arc archive
			arc.UnixEpoch = unixEpoch
			arc.Document = d

			storeNewsPost(string(b))
			xmlFile.Close()
		}

		// we have finished processing all XML files in archive so add
		// the key to the set so that we do not process it again
		storeNewsKey(unixEpoch)

		// decrement wait group counter, we have finished with this
		// file
		wg.Done()
	}
}

// unzipFile extracts the src archive into the dest directory
// WARNING: this method does not currently deal with embedded directories
// within the archive, it expects everything that is at the root of the archive
// to be a file
func unzipFile(src, dest string) {
	arc, err := zip.OpenReader(src)
	if err != nil {
		log.Fatal(err)
	}
	defer arc.Close()

	// make a dir to put all the xml files in
	os.MkdirAll(dest, 0755)

	extract := func(arcFile *zip.File) {
		// Open the archive file
		rc, err := arcFile.Open()
		if err != nil {
			log.Fatal(err)
		}
		defer rc.Close()

		// create a file to put the contents in
		path := filepath.Join(dest, arcFile.Name)
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, arcFile.Mode())
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		// copy the contents to the new file
		_, err = io.Copy(f, rc)
		if err != nil {
			log.Fatal(err)
		}
	}

	// for each file in the archive, extract it
	for _, f := range arc.File {
		extract(f)
	}
}

func getXMLFileNames(dir string) []string {
	fileList := []string{}
	filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		fileList = append(fileList, path)
		return nil
	})
	return fileList
}
