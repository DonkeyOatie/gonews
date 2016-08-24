package main

import (
	"fmt"
)

// getFileNames gets the file names as a slice of strings from the index.html
// at nuviNewsURL
func getFileNames() []string {
	var files []string
	files = append(files, "test one")
	files = append(files, "test two")
	files = append(files, "test three")
	files = append(files, "test four")
	files = append(files, "test five")
	files = append(files, "test six")
	files = append(files, "test seven")
	files = append(files, "test eight")
	files = append(files, "test nine")
	return files
}

// getFiles iterates over all file names in the list, and if it is a new file
// starts a routine to download it
func getFiles(fileNames []string, c chan string) {
	// get a file name from the channel
	for i, file := range fileNames {
		wg.Add(1)
		go getFile(file, c)
		// only start 4 routines for fetching files at any one time
		if i > 0 && i % 4 == 0 {
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
func getFile(fileName string, c chan string) {
	// get the file if the key is not in redis already

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
