package utils

// HOARD (c) 2018  Nicholas Albright and Deep End Research
// Core
import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

// External
import (
	"github.com/go-redis/redis"         // go get github.com/go-redis/redis
	"github.com/seiflotfy/cuckoofilter" // go get github.com/seiflotfy/cuckoofilter
)

func timestamp() (string, int64) {
	// Return our formatted time stamp
	ts := strings.Split(time.Now().Format(time.RFC3339), "T") // Split Date & Time
	cal := strings.Replace(ts[0], "-", "", -1)                // Strip dashes from Date
	hr := strings.Replace(ts[1][:5], ":", "", -1)             // Strip colon from time.
	st := int64(time.Now().Unix())
	return fmt.Sprintf("%s_%s", cal, hr), st // Return formated:  YYYYMMDD_HHMM, epochtime.
}

func writeFilter(filename string, cf *cuckoofilter.CuckooFilter) {
	fh, _ := os.Create("./sketches/" + filename + ".cf") // TODO: Need to make sure the directory is created.
	defer fh.Close()
	outstream := bufio.NewWriter(fh)
	outstream.Write(cf.Encode())
	outstream.Flush()
	return
}

func BuildSketch(tqname string, tserver string, dbID int) {
	rdb := redis.NewClient(&redis.Options{Addr: tserver, DB: dbID})
	for {
		start_ts, epoch := timestamp()
		cf := cuckoofilter.NewCuckooFilter(100000 + 2000) // Build specific sized filter with room for growth
		for i := 0; i < 100000; {                         // Will write at 100,000 unique observables or at a time interval, whichever happens first.
			msg, err := rdb.RPop(tqname).Result()
			if len(msg) > 0 && err == nil {
				x := cf.InsertUnique([]byte(strings.ToLower(msg))) // INSERT UNIQUE does all the hard work for us.
				if x {
					// fmt.Println("Just added another filter, now at: ", i) // DEBUG
					i++ // Increment our counter ONLY if we added one to the queue.
				}
			} else {
				time.Sleep(5 * time.Second) // Wait a few seconds to add more data to Queue.
			}
			if (int64(time.Now().Unix()) - epoch) >= 7200 { // 2 hours
				break // We need to write our sketches.
			}
		}
		// fmt.Println("Writing Sketches...") // DEBUG.
		fts, _ := timestamp()
		writeFilter("HOARD_"+start_ts+"-"+fts, cf) // We've hit prescribed observables, rotate.
	}
}
