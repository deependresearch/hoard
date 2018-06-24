package main

/*
HOARD (c) 2018  Nicholas Albright and Deep End Research

*/
import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// EXTERNAL LIBRARIES

import (
	"github.com/seiflotfy/cuckoofilter" // go get github.com/seiflotfy/cuckoofilter
)

func restoreDate(fname string) (ts []string) {
	startend := strings.Split(strings.Split(fname, "HOARD_")[1], "-")
	for item := range startend {
		var fd = []byte(startend[item])
		ts = append(ts, fmt.Sprintf("%s/%s/%s %s:%s", fd[0:4], fd[4:6], fd[6:8], fd[9:11], fd[11:13]))
	}
	return
}

func main() {
	intel := flag.String("i", "intel.txt", "Intelligence file.") // One observable per line. Exact matches (no wildcards)
	sketches := flag.String("s", "./sketches/", "Directory containing Sketches.")
	flag.Parse()

	filehandler_in, _ := os.Open(*intel)
	defer filehandler_in.Close()
	var IOCList []string

	readfile := bufio.NewScanner(filehandler_in) // Grab the IOC's from a text file, one per line.
	for readfile.Scan() {
		IOCList = append(IOCList, readfile.Text()) // Append to a slice (array) for review.
	}

	allsketches, _ := filepath.Glob(*sketches + "/*")
	for i := range allsketches {
		fstream, err := ioutil.ReadFile(allsketches[i])
		if err != nil {
			fmt.Println("Error Detected:", err)
			os.Exit(1)
		}
		cf, _ := cuckoofilter.Decode(fstream)
		for item := range IOCList {
			result := cf.Lookup([]byte(IOCList[item]))
			if result == true {
				ts := restoreDate(allsketches[i])
				fmt.Printf("[WARNING] IOC *%s* might have been observed between %s and %s.\n", IOCList[item], ts[0], ts[1])
			}
		}
	}
}
