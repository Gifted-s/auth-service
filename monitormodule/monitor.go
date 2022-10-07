package monitormodule

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	//"golang.org/x/crypto/tea"
	"gopkg.in/yaml.v2"
)

const endpoint string = "http://localhost:9090/metrics"
const outputEndpoint string = "http://localhost:9090/checkRoutine"
const metricConfFile string = "./monitormodule/config.yaml"

type conifgMap struct {
	Metadataregion     string
	Metadatapipelineid int
	Signin             string
	Signup             string
}

type transformationOutput struct {
	Time          time.Time
	Pipelineid    int
	Region        string
	MetricDetails []string
}

// constantly ping specified services for metrics
func dataAggregator(log *zap.Logger, key string, trackData conifgMap, aggregatorChan chan<- []string) error {
	// always check every 5 seconds
	fetchInterval := time.NewTicker(time.Second * 5)
	for range fetchInterval.C {
		// send request to service
		resp, err := http.Get(endpoint)

		if err != nil {
			log.Error("Error on fetching data", zap.Error(err))
			return err
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Error("Error", zap.Error(err))
			return err
		}

		contents := string(body)
		aggregate := []string{}

		// i have used this to get value dynamically from a struct
		reflection := reflect.ValueOf(trackData)
		reflectionField := reflect.Indirect(reflection).FieldByName(key)
		// get the keys we want to fetch
		allKeys := strings.Split(reflectionField.String(), ",")
		for _, metricValue := range strings.Split(contents, "\n") {
			for _, metricKey := range allKeys {
				// check if the given line has the metric we are looking for
				if strings.Contains(metricValue, metricKey) {
					// ignore the first character
					if metricValue[0] != '#' {
						aggregate = append(aggregate, metricValue)
					}
				}
			}
		}
		aggregatorChan <- aggregate
	}
	return nil
}

// dataTransformer reads from aggregate channel and transorms into a struct. This is output to a unique channel for this goroutine only.
func dataTransform(log *zap.Logger, routineId int, aggregatorChan <-chan []string, transformChan chan<- string, region string, pipelineId int) error {
	for msg := range aggregatorChan {
		tr := &transformationOutput{
			Time:          time.Now(),
			Pipelineid:    pipelineId,
			MetricDetails: msg,
			Region:        region,
		}
		op, err := json.Marshal(tr)
		if err != nil {
			log.Error("Error in marshalling data to json string", zap.Int("routine id", routineId), zap.Error(err))
			return err
		}

		transformChan <- string(op)
	}
	return nil
}

// dataTransportation send a POST request to our endpoint which is on our server only
func dataTransport(log *zap.Logger, transChan <-chan string, successChan chan<- string) error {

	for val := range transChan {
		resp, err := http.Post(outputEndpoint, "application/json", bytes.NewBuffer([]byte(val)))
		if err != nil {
			log.Error("Error", zap.Error(err))
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Error("Error", zap.Error(err))
			return err
		}

		successChan <- string(body)
	}
	return nil
}

// simple function that prints the channels value available. It takes in the Waitgroup
func showDemo(log *zap.Logger, finalChan <-chan string, wg *sync.WaitGroup) {
	log.Info("final Chan ", zap.String("message", <-finalChan))
	wg.Done()
}

// MonitorBinder binds all the goroutines and combines the output
func MonitorBinder(log *zap.Logger) error {
	content, err := ioutil.ReadFile(metricConfFile)
	if err != nil {
		return err
	}
	metricEndpoints := conifgMap{}
	yaml.Unmarshal(content, metricEndpoints)

	aggregatorChan := make(chan []string)
	go dataAggregator(log, "Signin", metricEndpoints, aggregatorChan)
	go dataAggregator(log, "Signup", metricEndpoints, aggregatorChan)

	transformChan1 := make(chan string)
	transformChan2 := make(chan string)

	go dataTransform(log, 1, aggregatorChan, transformChan1, metricEndpoints.Metadataregion, metricEndpoints.Metadatapipelineid)
	go dataTransform(log, 2, aggregatorChan, transformChan2, metricEndpoints.Metadataregion, metricEndpoints.Metadatapipelineid)

	transChannel := make(chan string)

	go func() {
		for {
			select {
			case msg1 := <-transformChan1:
				transChannel <- msg1
			case msg2 := <-transformChan2:
				transChannel <- msg2
			}
		}
	}()

	successChan := make(chan string)
	// Run on two goroutines to make process fast
	go dataTransport(log, transChannel, successChan)
	go dataTransport(log, transChannel, successChan)

	var wg sync.WaitGroup
	for i := 0; i <= 5; i++ {
		wg.Add(1)
		showDemo(log, successChan, &wg)
	}

	// this goroutine only runs after the first 5 responses have been printed
	go func() {
		wg.Wait()
		log.Info("=========== WE ARE DONE FOR THE DEMO ===========")
	}()
	return nil
}
