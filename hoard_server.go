package main

/*
HOARD (c) 2018  Nicholas Albright and Deep End Research

This app will pull data off a Redis Queue, extract IOCs and write to a Cuckoo Filter.
*/

import (
	"flag"
	"fmt"
	"hoard/utils"
	"reflect"
	"runtime"
	"sync"
	"time"
)

// This is the fastest way I know of to ignore internal IPs and such.
// Note, \d is faster than [0-9]. \d\d?\d? is faster than [0-9]{1,3}.
const ignore = `((0|10|127|224|239|255)\.\d\d?\d?\.\d\d?\d?\.\d\d?\d?)|192\.168\.\d\d?\d?\.\d\d?\d?|169\.254\.\d\d?\d?\.\d\d?\d?|((172\.1[6-9]\.)|(172\.2\d\.)|(172\.3[0-1]\.))\d\d?\d?\.\d\d?\d?`

// Predefined Configuration.
// should probably to a struct type for this and make it look better.  // TODO

var regex string
var server string = "127.0.0.1"     // Run Redis Locally.
var port string = "6379"            // Default Redis Port
var DB int = 0                      // For Redis PUSH/POP it appears DB Zero us used by Default. .
var logQ string = "input_queue"     // Our Input Sketches.
var sketchQ string = "sketch_queue" // Our queues for Sketches.
var eventType string = "JSON"       // Log or JSON
var wg sync.WaitGroup

func reassignConfig(conf map[string]interface{}) { // Re-Assign our config vars based upon the file we receive.

	if conf["redis_ip"] != nil {
		server = conf["redis_ip"].(string)
	}
	if conf["redis_port"] != nil {
		port = conf["redis_port"].(string)
	}
	if conf["log_queue"] != nil {
		logQ = conf["log_queue"].(string)
	}
	if conf["sketch_queue"] != nil {
		sketchQ = conf["sketch_queue"].(string)
	}
	if conf["event_type"] != nil {
		eventType = conf["event_type"].(string)
	}
	if conf["regex"] != nil {
		sl := reflect.ValueOf(conf["regex"]) // Iterate through the regex's in the json config.
		for i := 0; i < sl.Len(); i++ {
			regex += fmt.Sprintf("%s", sl.Index(i))
			if i < sl.Len()-1 { // Apply our "OR" value.
				regex += "|"
			}
		}
	}
}

func main() {

	defaultConfig := flag.String("c", "hoard_config.json", "Use different config file.")
	flag.Parse()

	reassignConfig(utils.ParseConfig(*defaultConfig)) // This function reassigns vars after parsing config file.

	fmt.Printf("Starting %d Queue Readers...", runtime.NumCPU()*2)

	for i := 1; i < (runtime.NumCPU() * 2); i++ { // Two Concurrent Threads per CPU (approx 3k EPS per thread, max best at 1/core)
		wg.Add(1)
		time.Sleep(500 * time.Millisecond) // Give a half second delay between queue reader starts.
		go utils.ReadQueue(regex, ignore, logQ, sketchQ, eventType, server+":"+port, DB)
	}
	wg.Add(1)
	fmt.Println("Launching SketchBuilder.")
	go utils.BuildSketch(sketchQ, server+":"+port, DB)
	wg.Wait()
}
