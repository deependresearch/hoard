package utils

// HOARD (c) 2018  Nicholas Albright and Deep End Research
import (
	"encoding/json"
	"regexp"
	"strings"
	"time"
)

// External Libs:
import (
	"github.com/go-redis/redis"
)

func ParseJSON(eve string) []string {
	/*
		Average processing speed with 4 procs = 12,000 EPS.
	*/
	var jsonData map[string]interface{}
	var dataret []string
	json.Unmarshal([]byte(eve), &jsonData)

	if jsonData["event_type"].(string) == "dns" {
		if jsonData["rrname"] == nil || strings.HasSuffix(jsonData["rrname"].(string), ".local") {
			return dataret // We don't want to store LOCAL servers.
		}
		dataret = append(dataret, jsonData["rrname"].(string))
		dataret = append(dataret, jsonData["rdata"].(string))
		return dataret
	}
	if jsonData["event_type"].(string) == "http" {
		hostname := jsonData["http"].(map[string]interface{})["hostname"]
		if hostname == nil || jsonData["dest_ip"] == nil {
			return dataret
		}
		dataret = append(dataret, hostname.(string))
		dataret = append(dataret, jsonData["dest_ip"].(string)) // in case we're calling to C2.
		dataret = append(dataret, jsonData["src_ip"].(string))  // in case we're the victims of an inbound attack.
		return dataret
	}
	return dataret
}

func ReadQueue(regex string, ignore string, tqname string, rqname string, eventtype string, server string, dbID int) {
	r := regexp.MustCompile(regex)
	d := regexp.MustCompile(ignore)
	rdb := redis.NewClient(&redis.Options{Addr: server, DB: dbID})
	ddb := redis.NewClient(&redis.Options{Addr: server, DB: dbID})

	for {
		msg, err := rdb.RPop(tqname).Result()

		if len(msg) > 0 && err == nil {
			if eventtype == "regex" || eventtype == "log" {
				matches := r.FindAllString(msg, -1)
				for i := range matches {
					fcheck := d.FindString(matches[i])
					if len(fcheck) < 1 || fcheck == "" { // No Match in the ignore, add to our queue.
						ddb.LPush(rqname, matches[i])
					}
				}
			} else {
				switch {
				case strings.Contains(msg, "NXDOMAIN") || strings.Contains(msg, `"type":"query"`):
					break // Don't continue to process if we can't do anything with the data.
				default:
					events := ParseJSON(msg)
					for i := range events {
						fcheck := d.FindString(events[i])
						if len(fcheck) < 1 || fcheck == "" { // No Match in the ignore, add to our queue.
							ddb.LPush(rqname, events[i])
						}
					}
				}
			}
		} else {
			time.Sleep(5 * time.Second) // Wait a few seconds to add more data to Queue.
		}
	}
	return
}
