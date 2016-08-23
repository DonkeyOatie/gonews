package main

import "fmt"

// getFileNames gets the file names as a slice of strings from the index.html
// at nuviNewsURL
func getFileNames() []string {
	var files []string
	files = append(files, "test one")
	files = append(files, "test two")
	return files
}

// getFiles iterates over all file names in the list, and if it is a new file
// starts a routine to download it
func getFiles(fileNames []string, c chan string) {
	defer wg.Done()
	// get a file name from the channel
	for _, file := range fileNames {
		getFile(file, c)
	}
	close(c)
}

// getFile downloads the file with name fileName and once downloaded, puts the
// name in the channel so processing can begin by whatever is listening to the
// channel
func getFile(fileName string, c chan string) {

	// get the file

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
	}
}
