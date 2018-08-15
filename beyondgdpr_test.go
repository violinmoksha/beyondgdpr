package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"runtime"
	"sync"
	"testing"
)

const baseURL string = "http://localhost:8080"
const encryptEndpoint string = "/encryptPlaintext"
const decryptEndpoint string = "/decryptCiphertext"

var testWg sync.WaitGroup

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func TestEncryptBadRequest(t *testing.T) {
	tests := []struct {
		plaintext string
		userkey   string
	}{
		{
			plaintext: "Hello, World!",
			userkey:   "+YbX43O5PU/o1bBlRoFh1pZTbluSzABjuxriVo3e+Bk=", // random key example
		},
	}

	for _, tt := range tests {
		// this POST is also incorrectly formed ...ditto
		response, err := http.PostForm(baseURL+encryptEndpoint, url.Values{
			"plaintext": {tt.plaintext},
			"userkey":   {tt.userkey},
		})
		if err != nil {
			t.Fatal(err)
		}

		defer response.Body.Close()

		if response.Status != "400 Bad Request" {
			t.Errorf("Should have returned 400 Bad Request")
		}
	}
}

func TestDecryptBadRequest(t *testing.T) {
	tests := []struct {
		ciphertext string
		userkey    string
	}{
		{
			ciphertext: "4UUcJTQgZm06xfBxOVQ8SaewnqvZsAFhgT-__hqxFlvbRgzUp4rKAYw=",
			userkey:    "+YbX43O5PU/o1bBlRoFh1pZTbluSzABjuxriVo3e+Bk=",
		},
	}

	for _, tt := range tests {
		// this POST is incorrectly formed thus does not pass a [44]rune aka [32]byte userkey
		// and should therefore cause server to issue 400 without crashing it
		response, err := http.PostForm(baseURL+decryptEndpoint, url.Values{
			"ciphertext": {tt.ciphertext},
			"userkey":    {tt.userkey},
		})
		if err != nil {
			t.Fatal(err)
		}

		defer response.Body.Close()

		if response.Status != "400 Bad Request" {
			t.Errorf("Should have returned 400 Bad Request")
		}
	}
}

func TestEncryptDecryptRoundtrip(t *testing.T) {
	tests := []struct {
		plaintext string
		userkey   string
	}{
		{
			plaintext: "Hello, World! Here is an even longer plaintext which is really, really long...",
			userkey:   "+YbX43O5PU/o1bBlRoFh1pZTbluSzABjuxriVo3e+Bk=",
		},
	}

	for _, tt := range tests {
		var jsonStr = []byte(`{"plaintext":"` + tt.plaintext + `","userkey":"` + tt.userkey + `"}`)
		req, err := http.NewRequest("POST", baseURL+encryptEndpoint, bytes.NewBuffer(jsonStr))
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Errorf("Should have gotten client response: %s", err.Error())
		}
		defer resp.Body.Close()

		if resp.Status != "200 OK" {
			t.Errorf("Should have returned 200 OK from encryptEndpoint")
		}

		// now get that ciphertext out
		body, _ := ioutil.ReadAll(resp.Body)
		c := make(map[string]interface{})
		e := json.Unmarshal(body, &c)
		if e != nil {
			t.Errorf("Should have been able to unmarshall json from encryptEndpoint: %s", e.Error())
		}
		k := make([]string, len(c))
		i := 0
		for s := range c {
			k[i] = s
			i++
		}
		if k[0] != "ciphertext" {
			t.Errorf("Should have gotten some ciphertext back from encryptEndpoint")
		}

		var objmap map[string]*json.RawMessage
		e = json.Unmarshal(body, &objmap)
		if e != nil {
			t.Errorf("Should have been able to unmarshall json from encryptEndpoint: %s", e.Error())
		}
		var ourCiphertext string
		e = json.Unmarshal(*objmap["ciphertext"], &ourCiphertext)
		if e != nil {
			t.Errorf("Should have been able to pull ciphertext out of encryptEndpoint response: %s", e.Error())
		}

		// and go back to decryptEndpoint with it
		jsonStr = []byte(`{"ciphertext":"` + ourCiphertext + `","userkey":"` + tt.userkey + `"}`)
		req, err = http.NewRequest("POST", baseURL+decryptEndpoint, bytes.NewBuffer(jsonStr))
		req.Header.Set("Content-Type", "application/json")

		client = &http.Client{}
		resp, err = client.Do(req)
		if err != nil {
			t.Errorf("%s", err.Error())
		}
		defer resp.Body.Close()

		if resp.Status != "200 OK" {
			t.Errorf("Should have returned 200 OK from decryptEndpoint")
		}

		// and finally, pull plaintext out which should match initial plaintext
		body, _ = ioutil.ReadAll(resp.Body)
		var decryptObjmap map[string]*json.RawMessage
		e = json.Unmarshal(body, &decryptObjmap)
		if e != nil {
			t.Errorf("Should have been able to unmarshall json from decryptEndpoint: %s", e.Error())
		}
		var ourPlaintext string
		e = json.Unmarshal(*decryptObjmap["plaintext"], &ourPlaintext)
		if e != nil {
			t.Errorf("Should have been able to pull plaintext out of decryptEndpoint response: %s", e.Error())
		}

		if ourPlaintext != tt.plaintext {
			t.Errorf("Should be getting original plaintext back after encryptEndpoint+decryptEndpoint")
		}
	}
}

func BenchmarkLoads(b *testing.B) {
	for i := 0; i < b.N; i++ {
		testWg.Add(1)
		go encryptDecryptRoundtripForBench()
		testWg.Wait()
	}
}

func encryptDecryptRoundtripForBench() {
	tests := []struct {
		plaintext string
		userkey   string
	}{
		{
			plaintext: "Hello, World! Here is an even longer plaintext which is really, really long...",
			userkey:   "+YbX43O5PU/o1bBlRoFh1pZTbluSzABjuxriVo3e+Bk=",
		},
	}

	for _, tt := range tests {
		var jsonStr = []byte(`{"plaintext":"` + tt.plaintext + `","userkey":"` + tt.userkey + `"}`)
		req, _ := http.NewRequest("POST", baseURL+encryptEndpoint, bytes.NewBuffer(jsonStr))
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
		}
		defer resp.Body.Close()

		// now get that ciphertext out
		body, _ := ioutil.ReadAll(resp.Body)
		var objmap map[string]*json.RawMessage
		_ = json.Unmarshal(body, &objmap)
		var ourCiphertext string
		_ = json.Unmarshal(*objmap["ciphertext"], &ourCiphertext)

		// and go back to decryptEndpoint with it
		jsonStr = []byte(`{"ciphertext":"` + ourCiphertext + `","userkey":"` + tt.userkey + `"}`)
		req, _ = http.NewRequest("POST", baseURL+decryptEndpoint, bytes.NewBuffer(jsonStr))
		req.Header.Set("Content-Type", "application/json")

		client = &http.Client{}
		resp, err = client.Do(req)
		if err != nil {
		}
		defer resp.Body.Close()

		// and finally, pull plaintext out which should match initial plaintext
		body, _ = ioutil.ReadAll(resp.Body)
		var decryptObjmap map[string]*json.RawMessage
		_ = json.Unmarshal(body, &decryptObjmap)
		var ourPlaintext string
		_ = json.Unmarshal(*decryptObjmap["plaintext"], &ourPlaintext)
	}

	testWg.Done()
}
