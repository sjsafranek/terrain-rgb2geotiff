package main

import (
	"crypto/tls"
	"net/http"
	"time"
)

var tr = &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}

var client = &http.Client{
	Timeout:   time.Second * 60,
	Transport: tr,
}
