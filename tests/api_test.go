package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/eensymachines-in/patio-web/devices"
	"github.com/stretchr/testify/assert"
)

var (
	config_200OK = []map[string]interface{}{
		{"config": 2, "tickat": "10:00", "pulsegap": 50, "interval": 100},
		{"config": 1, "tickat": "10:00", "pulsegap": 50, "interval": 100},
		{"config": 0, "tickat": "10:00", "pulsegap": 50, "interval": 100},
		{"config": 3, "tickat": "10:00", "pulsegap": 50, "interval": 100},
	}
	data_200OK = []devices.Device{
		{
			Location: "18.41828900932213, 73.76933368232831",
			Mac:      devices.MacID("45-36-17-E3-1C-70"),
			Name:     "Aquaponics pump control-I0",
			Make:     "Raspberry Pi 0w 512MB, 1GRM",
			Users: []string{
				"jionesco0@globo.com",
				"koreilly1@wufoo.com",
				"cdomoni2@ycombinator.com",
			},
		},
		{
			Location: "", // with no location , should be ok
			Mac:      devices.MacID("45-36-17-E3-1C-60"),
			Name:     "Aquaponics pump control-II", // with the same name, should be ok ?
			Make:     "Raspberry Pi 0w 512MB, 1GRM",
			Users: []string{
				"jionesco0@globo.com",
				"koreilly1@wufoo.com",
				"cdomoni2@ycombinator.com",
			},
		},
		{
			Location: "", // with no location , should be ok
			Mac:      devices.MacID("45-36-17-E3-1C-55"),
			Name:     "Aquaponics pump control-II", // with the same name, should be ok ?, but different MAc
			Make:     "Raspberry Pi 0w 512MB, 1GRM",
			Users: []string{
				"jionesco0@globo.com",
				"koreilly1@wufoo.com",
				"cdomoni2@ycombinator.com",
			},
		},
	}
)

func TestPatchDeviceConfig(t *testing.T) {
	cl := &http.Client{
		Timeout: 5 * time.Second,
	}
	device := data_200OK[0]
	for _, d := range config_200OK {
		url := fmt.Sprintf("http://localhost:8081/api/devices/%s/config", device.Mac)
		byt, _ := json.Marshal(d)
		req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(byt))
		assert.Nil(t, err, "Unexpected error when creating a new get http request")
		resp, err := cl.Do(req)

		assert.Nil(t, err, "Check your netwrok connection , request did not go thru")
		assert.Equal(t, 200, resp.StatusCode, "Unexpected status code from server")
	}
}

func TestGetDeviceConfig(t *testing.T) {
	cl := &http.Client{
		Timeout: 5 * time.Second,
	}
	for _, d := range data_200OK {
		url := fmt.Sprintf("http://localhost:8081/api/devices/%s/config", d.Mac)
		req, err := http.NewRequest("GET", url, nil)
		assert.Nil(t, err, "Unexpected error when creating a new get http request")
		resp, err := cl.Do(req)

		assert.Nil(t, err, "Check your netwrok connection , request did not go thru")
		assert.Equal(t, 200, resp.StatusCode, "Unexpected status code from server")
		if resp.StatusCode == 200 {
			config := map[string]interface{}{}
			byt, _ := io.ReadAll(resp.Body)
			err := json.Unmarshal(byt, &config)
			assert.Nil(t, err, "Unexpected error when reading response body")
			t.Log(config)
		}

	}
}

func TestDelDevice(t *testing.T) {
	/* TEST: cleanng up the test  */

	cl := &http.Client{
		Timeout: 5 * time.Second,
	}
	for _, d := range data_200OK {
		url := fmt.Sprintf("http://localhost:8081/api/devices/%s", d.Mac)
		req, err := http.NewRequest("DELETE", url, nil)
		assert.Nil(t, err, "Unexpected error when creating a del http request")
		resp, err := cl.Do(req)

		assert.Nil(t, err, "Check your netwrok connection , request did not go thru")
		assert.Equal(t, 200, resp.StatusCode, "Unexpected status code from server")
	}
}

func TestAddGetDevice(t *testing.T) {
	url := "http://localhost:8081/api/devices/"

	cl := &http.Client{
		Timeout: 5 * time.Second,
	}
	for _, d := range data_200OK {
		byt, err := json.Marshal(d)
		assert.Nil(t, err, "Unexpected error when marshalling data to json string")

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(byt))
		assert.Nil(t, err, "Unexpected error when creating a new post http request")
		req.Header.Set("Content-Type", "application/json; charset=utf-8") // cannot miss setting this
		resp, err := cl.Do(req)

		assert.Nil(t, err, "Check your netwrok connection , request did not go thru")
		assert.Equal(t, 200, resp.StatusCode, "Unexpected status code from server")
	}
}

func TestPutDeviceConfig(t *testing.T) {
	OK_data := []map[string]interface{}{
		{"config": 0, "tickat": "12:10", "pulsegap": 50, "interval": 100},
		{"config": 1, "tickat": "12:10", "pulsegap": 50, "interval": 100},
		{"config": 2, "tickat": "12:10", "pulsegap": 50, "interval": 100},
		{"config": 3, "tickat": "12:10", "pulsegap": 50, "interval": 100},
	}
	BAD_data := []map[string]interface{}{
		{"config": 10, "tickat": "12:10", "pulsegap": 50, "interval": 100}, // config range overflow
		{"config": 2, "tickat": "12:10", "pulsegap": 100, "interval": 50},  //  pulsegap > interval for config 2
		{"config": 1, "tickat": "", "pulsegap": 50, "interval": 100},       // tickat empty for config 1
		{"config": 3, "tickat": "", "pulsegap": 50, "interval": 100},       // tickat empty for config 3
	}
	url := "http://localhost:8081/api/devices/705b200f-059b-4630-bf3d-5d55c3a4a9dc/config"

	cl := &http.Client{
		Timeout: 10 * time.Second,
	}
	for _, d := range OK_data {
		byt, err := json.Marshal(d)
		assert.Nil(t, err, "Unexpected error when marshalling data for payload")

		req, err := http.NewRequest("PUT", url, bytes.NewBuffer(byt))
		assert.Nil(t, err, "Unexpected err when making the request")
		req.Header.Set("Content-Type", "application/json; charset=utf-8") // this is important

		resp, err := cl.Do(req)

		assert.Nil(t, err, "Unexpected sending the request, are you connected to the server?")
		assert.Equal(t, 200, resp.StatusCode, "Unexpected status code for a device that is registered")
	}
	for _, d := range BAD_data {
		byt, err := json.Marshal(d)
		assert.Nil(t, err, "Unexpected error when marshalling data for payload")

		req, err := http.NewRequest("PUT", url, bytes.NewBuffer(byt))
		assert.Nil(t, err, "Unexpected err when making the request")
		req.Header.Set("Content-Type", "application/json; charset=utf-8") // this is important

		resp, err := cl.Do(req)

		assert.Nil(t, err, "Unexpected sending the request, are you connected to the server?")
		assert.Equal(t, 400, resp.StatusCode, "Unexpected status code for a device that is registered")
	}
}
