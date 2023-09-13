package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type PrometheusResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric struct {
				Instance string `json:"instance"`
			} `json:"metric"`
			Values [][]interface{} `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

func main() {

	metrics := []string{
		"ping_latency_millonseconds",
		"ping_packageDrop_rate",
		"ping_target_success",
		"ping_rttavg",
		"ping_rttmin",
		"ping_rttmax",
		"ping_rttmdev",
	}
	client := &http.Client{
		Timeout: time.Second * 120,
	}

	fileNames := []string{
		"data/ping_latency.json",
		"data/ping_packageDrop_rate.json",
		"data/ping_target_success.json",
		"data/ping_rttavg.json",
		"data/ping_rttmin.json",
		"data/ping_rttmax.json",
		"data/ping_rttmdev.json",
	}

	for i, metric := range metrics {

		url := fmt.Sprintf("http://localhost:9090/api/v1/query_range?query=%s&start=2023-08-01T00:00:00.781Z&end=2023-08-03T00:00:00.781Z&step=1m", metric)

		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			fmt.Printf("[ERROR] Failed to create request for %q: %v\n", metric, err)
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("[ERROR] Failed to get response for %q: %v\n", metric, err)
			continue
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("[ERROR] Failed to read response body for %q: %v\n", metric, err)
			continue
		}

		fileName := fileNames[i]

		err = ioutil.WriteFile(fileName, body, 0644)
		if err != nil {
			fmt.Printf("[ERROR] Failed to write response to file %q: %v\n", fileName, err)
			continue
		}

		fmt.Printf("Saved response for %q to %q.\n", metric, fileName)
	}

}
