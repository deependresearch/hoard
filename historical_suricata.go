package main

/*
Use this application to parse OLD suricata events.

Note - you'll need to ungzip them first if you compress on rotate

This is just a quick hack to get old data into the sketches
This shouldn't be used as the primary method of collecting data,
though it will likely work just fine in a pinch or as part of a
pre/post rotate script.
*/

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/seiflotfy/cuckoofilter"
	"os"
	"strings"
	"time"
)

func ParseJSON(eve string) ([]string, string) {
	var jsonData map[string]interface{}
	var dataret []string
	json.Unmarshal([]byte(eve), &jsonData)
	timestamp := jsonData["timestamp"].(string)
	if jsonData["event_type"].(string) == "dns" {
		if jsonData["rrname"] == nil || strings.HasSuffix(jsonData["rrname"].(string), ".local") {
			return dataret, timestamp // We don't want to store LOCAL servers.
		}
		dataret = append(dataret, jsonData["rrname"].(string))
		dataret = append(dataret, jsonData["rdata"].(string))
		return dataret, timestamp
	}
	if jsonData["event_type"].(string) == "http" {
		hostname := jsonData["http"].(map[string]interface{})["hostname"]
		if hostname == nil || jsonData["dest_ip"] == nil {
			return dataret, timestamp
		}
		dataret = append(dataret, hostname.(string))
		dataret = append(dataret, jsonData["dest_ip"].(string))
		dataret = append(dataret, jsonData["src_ip"].(string))
		return dataret, timestamp
	}
	return dataret, timestamp
}

func writeFilter(filename string, cf *cuckoofilter.CuckooFilter) {
	fh, _ := os.Create("./sketches/" + filename + ".cf")
	defer fh.Close()
	outstream := bufio.NewWriter(fh)
	outstream.Write(cf.Encode())
	outstream.Flush()
	return
}

func fixtime(timedatestamp string) string {
	// Return our formatted time stamp
	ts := strings.Split(timedatestamp, "T")       // Split Date & Time
	cal := strings.Replace(ts[0], "-", "", -1)    // Strip dashes from Date
	hr := strings.Replace(ts[1][:5], ":", "", -1) // Strip colon from time.
	return fmt.Sprintf("%s_%s", cal, hr)          // Return formated:  YYYYMMDD_HHMM
}

func main() {

	var evefile = flag.String("f", "./eve.json", "Path to eve.json file")
	flag.Parse()
	cf := cuckoofilter.NewCuckooFilter(1000000 + 2000)
	filehandler_in, _ := os.Open(*evefile) // User supplied valid file.
	readfile := bufio.NewScanner(filehandler_in)
	var starttime int64
	var initts string
	for readfile.Scan() {
		parse, ts := ParseJSON(readfile.Text())
		t, err := time.Parse("2006-01-02T15:04:05.999999999-0700", ts)
		if err != nil {
			fmt.Println(err)
		}
		if starttime == 0 {
			starttime = t.Unix()
			initts = ts
		}
		if (t.Unix() - starttime) >= 7200 {
			starttime = t.Unix()
			initts = ts
			fmt.Println("Writing filter...")
			writeFilter("HOARD_"+fixtime(initts)+"-"+fixtime(ts), cf)
			cf = cuckoofilter.NewCuckooFilter(1000000 + 2000)
		}
		for item := range parse {
			cf.InsertUnique([]byte(strings.ToLower(parse[item])))
		}
	}
}
