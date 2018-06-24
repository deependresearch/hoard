package utils

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"
)

func SplitConf(svalue string) (sdata string, db int, queue string) {
	a := strings.Split(":", svalue)
	sdata = string(a[0]) + ":" + string(a[1])
	db, _ = strconv.Atoi(a[2]) // Convert the ascii string provide in our config to an int.
	queue = string(a[3])
	return
}

func ParseConfig(configFile string) (jsonCfgData map[string]interface{}) {
	/* Parse JSON Config File */
	fname, err := os.Open(configFile)
	if err != nil {
		log.Fatal("Unable to open config file.")
	}
	defer fname.Close()
	decoder := json.NewDecoder(fname)
	err = decoder.Decode(&jsonCfgData)
	if err != nil {
		log.Fatal(err)
	}
	return
}
