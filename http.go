package torrent

import (
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"sort"
	"time"
)

type HttpCallOptions struct {
	ConnectionTimeout time.Duration
	ReadWriteTimeout time.Duration
}

func httpGetUrl(baseUrl string, parameters map[string]string) string {
	output := ""

	var keys []string
	for key, _ := range parameters {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		if output != "" { output += "&" }
		output += key + "=" + url.QueryEscape(parameters[key])
	}
	if output != "" { output = "?" + output }
	return baseUrl + output
}

func CustomDialer(connectionTimeout time.Duration, readWriteTimeout time.Duration) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, connectionTimeout)
		if err != nil {
			return nil, err
		}
		conn.SetDeadline(time.Now().Add(readWriteTimeout))
		return conn, nil
	}
}

func NewHttpCallOptions() *HttpCallOptions {
	o := new(HttpCallOptions)
	o.ConnectionTimeout = 10 * time.Second
	o.ReadWriteTimeout = 10 * time.Second
	return o
}

func httpClient(options *HttpCallOptions) *http.Client {
	return &http.Client {
		Transport: &http.Transport {
			Dial: CustomDialer(options.ConnectionTimeout, options.ReadWriteTimeout),
	}}
}

func httpGet(url string, options *HttpCallOptions) ([]byte, error) {
	client := httpClient(options)
	response, err := client.Get(url)
	if err != nil { return nil, err }
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil { return nil, err }
	return body, nil
}