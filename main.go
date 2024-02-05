package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type URL struct {
	URL string
}
type CommonHdr struct {
	First bool
	Last  bool
	Count int
}

type Topo struct {
	URL
	Id     string `json:"id"`
	Name   string `json:"name"`
	TopoId string `json:"topoid"`
}

type Device struct {
	URL
	Id       string `json:"id"`
	Name     string `json:"name"`
	DeviceId string `json:"deviceid"`
}

type Topos struct {
	CommonHdr
	URL
	Data []Topo `json:"data"`
}

type Devices struct {
	CommonHdr
	URL
	Data []Device `json:"data"`
}

type API interface {
	Topos | Devices | Device
	getURL() string
}

func (d Devices) getURL() string {
	return d.URL.URL
}

func (d Topos) getURL() string {
	return d.URL.URL
}

func (d Device) getURL() string {
	return d.URL.URL
}

func (d Topo) getURL() string {
	return d.URL.URL
}

func Fetch[T API](url string, ret *T) error {

	jsonBody := []byte(`{"client_message": "hello, server!"}`)
	bodyReader := bytes.NewReader(jsonBody)

	req, err := http.NewRequest(http.MethodGet, url, bodyReader)
	if err != nil {
		log.Printf("err: %s", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Printf("req err: %s", err)
		return err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	if res.StatusCode < 400 {
		err = json.Unmarshal(body, ret)
		if err != nil {
			log.Printf("err unmarshall: %s: %s", err, body)
			return err
		}
		return nil
	} else {
		log.Printf("Status code=%d, body=%s", res.StatusCode, body)
	}
	return nil
}

func worker(id int, jobs <-chan int, results chan<- string) {

	for j := range jobs {
		log.Printf("Worker %d started job %d", id, j)
		time.Sleep(time.Second)
		results <- fmt.Sprintf("Finish w %d job %d", id, j)
	}

}

func fetchWorker[T API](id int, jobs <-chan T, results chan<- string) {

	for j := range jobs {
		log.Printf("Worker %d started job URL=%s", id, j.getURL())
		time.Sleep(time.Second)
		results <- fmt.Sprintf("Finish w %d job %v", id, j)
	}

}

func prepareSubmit[T API](parm []T) {

	numJ := len(parm)
	jobs := make(chan T, numJ)
	res := make(chan string, numJ)

	log.Printf("Spawn workers...")
	for w := 1; w <= 3; w++ {
		go fetchWorkerT(w, jobs, res)
	}

	// Send jobs
	for j := 0; j < numJ; j++ {
		jobs <- T[j]
	}
	close(jobs)
	// Wait results
	log.Printf("Wait results")
	for a := 1; a <= numJ; a++ {
		log.Printf("Res: %s", <-res)
	}

}

func main() {
	const numJ = 5
	jobs := make(chan Device, numJ)
	res := make(chan string, numJ)

	// Spawn workers
	log.Printf("Spawn workers...")
	for w := 1; w <= 3; w++ {
		go fetchWorker(w, jobs, res)
	}

	// Send jobs
	log.Printf("Start work")
	for j := 1; j <= numJ; j++ {
		d := Device{
			URL: URL{URL: "http://devices"},
		}
		jobs <- d
	}
	close(jobs)
	// Wait results
	log.Printf("Wait results")
	for a := 1; a <= numJ; a++ {
		log.Printf("Res: %s", <-res)
	}

}

func fetchW() {
	const numJ = 5
	jobs := make(chan int, numJ)
	res := make(chan string, numJ)

	// Spawn workers
	log.Printf("Spawn workers...")
	for w := 1; w <= 3; w++ {
		go worker(w, jobs, res)
	}

	// Send jobs
	log.Printf("Start work")
	for j := 1; j <= numJ; j++ {
		jobs <- j
	}
	close(jobs)
	// Wait results
	log.Printf("Wait results")
	for a := 1; a <= numJ; a++ {
		log.Printf("Res: %s", <-res)
	}

}
func fetchThings() {
	ds := Devices{}
	ts := Topos{}
	d := Device{}

	Fetch("http://localhost:8080/devices", &ds)
	for _, de := range ds.Data {
		log.Printf("Device ID=%s", de.DeviceId)
	}

	Fetch("http://localhost:8080/topos", &ts)
	for _, to := range ts.Data {
		log.Printf("Topo ID=%s", to.TopoId)
	}

	for i := 0; i < 10; i++ {
		Fetch(fmt.Sprintf("http://localhost:8080/devices/%d", i), &d)
		log.Printf("Device ID is *%s*", d.DeviceId)
	}

}
