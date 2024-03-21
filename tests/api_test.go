package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/eensymachines-in/patio/aquacfg"
	"github.com/stretchr/testify/assert"
)

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

func TestGetDeviceConfig(t *testing.T) {
	OK_data := []string{ // all the urls that are 200 ok
		"http://localhost:8081/api/devices/705b200f-059b-4630-bf3d-5d55c3a4a9dc/config",
	}
	NF_data := []string{ //i all the urls that are 400 not found
		"http://localhost:8081/api/devices/e71a85db-7797-4717-ae2e-a7c283c5093b/config",
		"http://localhost:8081/api/devices/d218ffb8-30bf-4cbf-9709-b431216cb438/config",
		"http://localhost:8081/api/devices/77003cca-2e4f-4207-8de6-010c28e27b24/config",
	}
	cl := &http.Client{
		Timeout: 10 * time.Second,
	}

	for _, d := range OK_data {
		req, err := http.NewRequest("GET", d, nil)
		assert.Nil(t, err, "Unexpected err when making the request")
		resp, err := cl.Do(req)
		assert.Nil(t, err, "Unexpected sending the request, are you connected to the server?")
		assert.Equal(t, 200, resp.StatusCode, "Unexpected status code for a device that is registered")
		byt, err := io.ReadAll(resp.Body)
		assert.Nil(t, err, "Unexpected err when reading the response payload")
		defer resp.Body.Close()
		appCfg := aquacfg.AppConfig{}
		err = json.Unmarshal(byt, &appCfg.Schedule)
		assert.Nil(t, err, "Unexpected err when unmarshalling the response payload")
		t.Log(appCfg)
	}
	for _, d := range NF_data {
		req, err := http.NewRequest("GET", d, nil)
		assert.Nil(t, err, "Unexpected err when making the request")
		resp, err := cl.Do(req)
		assert.Nil(t, err, "Unexpected sending the request, are you connected to the server?")
		assert.Equal(t, 404, resp.StatusCode, "Unexpected status code for a device that is registered")
	}
}
