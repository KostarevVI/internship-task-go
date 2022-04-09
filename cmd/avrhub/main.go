package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Item struct {
	Average     *float32 `json:"average"`
	IsCorrupted *bool    `json:"isCorrupted"`
}

// checkSensors continuously repeating the "erase-read-wait" cycle
func checkSensors(sNumber int, sDataColl *SensorDataCollection) {
	for {
		sDataColl.ResetFlags()
		for i := 0; i < sNumber; i++ {
			requestValue(i, sDataColl)
		}
		sDataColl.CalculateAverage()
		time.Sleep(30 * time.Second)
	}
}

// requestValue asks for value from sensor via HTTP
func requestValue(i int, sDataColl *SensorDataCollection) {
	go func() {
		sDataColl.Wg.Add(1)
		defer sDataColl.Wg.Done()

		url := "http://localhost:" + strconv.Itoa(8081+i)

		resp, err := http.Get(url)
		if err != nil {
			log.Fatalln(err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}
		result, _ := strconv.ParseInt(string(body), 10, 10)
		sDataColl.UpdateData(i, int(result), true)

		fmt.Println(url, ":", result)
		return
	}()
}

// main is main
// maybe it all would be working better with working pool or smth
func main() {
	sFlag := flag.Int("s", 4, "number of sensors to check")
	flag.Parse()
	sNumber := *sFlag
	if sNumber < 2 || sNumber > 10 {
		log.Println("Unexpected amount of sensors. Setting to default: 4")
		sNumber = 4
	}

	sDataColl := SensorDataCollection{}
	sDataColl.InitCollection(sNumber)

	go checkSensors(sNumber, &sDataColl)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if sDataColl.Average < 0 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		item := Item{
			&sDataColl.Average,
			&sDataColl.IsCorrupted,
		}

		sDataJson, _ := json.Marshal(item)
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(sDataJson)
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
