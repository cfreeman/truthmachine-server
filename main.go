/*
 * Copyright (c) Clinton Freeman 2018
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of this software and
 * associated documentation files (the "Software"), to deal in the Software without restriction,
 * including without limitation the rights to use, copy, modify, merge, publish, distribute,
 * sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all copies or
 * substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT
 * NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
 * NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
 * DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package main

import (
	"fmt"
	"github.com/hypebeast/go-osc/osc"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func pulse(heartRate chan int) {
	log.Println("Started the fully hectic Nissan Pulsar.")

	pulseLength := 1000
	start := time.Now()

	for {
		hr := <-heartRate
		if hr > 0 {
			pulseLength = 60000 / hr
		}

		if time.Now().Sub(start) > (time.Duration(pulseLength) * time.Millisecond) {
			// Broadcast the heartbeat.
			client := osc.NewClient("localhost", 53000)
			msg := osc.NewMessage("/cue/p/start")
			client.Send(msg)
			log.Println(msg.Address)

			start = time.Now()
		}
	}
}

func main() {
	log.Println("Starting TruthMachine v0.0.6")

	heartRate := make(chan int)
	go pulse(heartRate)

	addr := "localhost:8765"
	server := &osc.Server{Addr: addr}

	for _, endPoint := range []string{"/calibrate", "/interrogate", "/reset"} {
		server.Handle(endPoint, func(msg *osc.Message) {
			resp, err := http.Get("http://192.168.86.112/arduino/" + msg.Address)
			if err != nil {
				log.Println("Unable to contact theatrical polygraph")
			}

			log.Println(resp)
		})
	}
	go server.ListenAndServe()

	log.Println("Creating Qlab endpoint: '/cue/p/start'")
	http.HandleFunc("/h", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")

		f, err := strconv.ParseFloat(r.URL.Query()["v"][0], 32)
		if err != nil {
			log.Fatal("Unable to parse argument for '/h'.")
		}

		select {
		case heartRate <- int(f):
		default:
		}

	})

	for _, endPoint := range []string{"/l", "/r", "/g"} {
		log.Println("Creating HTTP endpoint: '" + endPoint + "'")
		http.HandleFunc(endPoint, func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "OK")

			client := osc.NewClient("localhost", 53000)

			msg := osc.NewMessage(strings.Split(r.URL.RequestURI(), "?")[0])

			f, err := strconv.ParseFloat(r.URL.Query()["v"][0], 32)
			if err != nil {
				log.Fatal("Unable to parse '" + endPoint + "' argument.")
			}

			log.Println(msg.Address)
			msg.Append(float32(f))
			client.Send(msg)
		})
	}

	log.Fatal(http.ListenAndServe(":8080", nil))
}
