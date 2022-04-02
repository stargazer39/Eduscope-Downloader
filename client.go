package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

type HttpClient struct {
	Client *http.Client
}

func NewHttpClient() *HttpClient {
	jar, _ := cookiejar.New(nil)

	client := &http.Client{
		Timeout: 30 * time.Second,
		Jar:     jar,
	}

	return &HttpClient{
		Client: client,
	}
}

func (this *HttpClient) GetJson(url string, s interface{}, query *url.Values) error {
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return err
	}

	if query != nil {
		req.URL.RawQuery = query.Encode()
	}

	resp, err := this.Client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	bytes, err := io.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	if err := json.Unmarshal(bytes, &s); err != nil {
		return err
	}

	return nil
}

func (this *HttpClient) GetJsonWithJsonParams(url string, s interface{}, r interface{}, query *url.Values) error {
	jBytes, bErr := json.Marshal(r)

	if bErr != nil {
		return bErr
	}

	bytesReader := bytes.NewReader(jBytes)

	req, err := http.NewRequest("GET", url, bytesReader)

	if err != nil {
		return err
	}

	if query != nil {
		req.URL.RawQuery = query.Encode()
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", "FBI")
	req.Header.Add("Content-Type", "application/json")

	resp, err := this.Client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	bytes, err := io.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	if err := json.Unmarshal(bytes, &s); err != nil {
		return err
	}

	return nil
}

func (this *HttpClient) GetString(url string) (error, string) {
	resp, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return err, ""
	}

	defer resp.Body.Close()

	bytes, err := io.ReadAll(resp.Body)

	if err != nil {
		return err, ""
	}

	return err, string(bytes)
}

func (this *HttpClient) PostForm(url string, val url.Values) (resp *http.Response, err error) {
	return this.Client.PostForm(url, val)
}
