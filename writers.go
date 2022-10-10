package log

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"time"
)

var client = &http.Client{
	Timeout: time.Second * 3,
	Transport: &http.Transport{
		TLSHandshakeTimeout:   time.Second,
		ResponseHeaderTimeout: time.Second,
		ExpectContinueTimeout: time.Second,
	},
}

type postWriter string

func (w postWriter) Write(p []byte) (n int, err error) {
	if w != "" {
		go func(url string, p []byte) {
			res, err := client.Post(url, "application/json", bytes.NewReader(p))

			if err != nil {
				log.Printf("internal/logger postWriter error: %s", err.Error())
			}

			if res == nil {
				log.Print("internal/logger postWriter nil client response")
				return
			}

			defer res.Body.Close()

			if res.StatusCode >= 300 {
				log.Printf("internal/loggger postwriter error: %s", res.Status)
			}
		}(string(w), p)
	}

	return fmt.Println(string(p))
}

type stdOutWriter struct{}

func (s stdOutWriter) Write(p []byte) (n int, err error) {
	return fmt.Println(string(p))
}
