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
	"math"
	"net/http"
	"strconv"
	"time"
)

func lerp(srcMin float64, srcMax float64, val float64, dstMin int, dstMax int) int {

	ratio := (math.Min(srcMax, math.Max(srcMin, val)) - srcMin) / (srcMax - srcMin)

	return int(ratio*float64(dstMax-dstMin)) + dstMin
}

func pulse(heartRate chan int) {
	log.Println("Started the fully hectic Nissan Pulsar.")

	pulseLength := 1000
	start := time.Now()

	for {
		select {
		case hr := <-heartRate:
			if hr > 0 {
				pulseLength = 60000 / hr
			}
		default:
		}

		if time.Now().Sub(start) > (time.Duration(pulseLength) * time.Millisecond) {
			// Broadcast the heartbeat.
			client := osc.NewClient("localhost", 53000)
			msg := osc.NewMessage("/cue/p/start")
			client.Send(msg)
			log.Println(msg.Address)

			start = time.Now()
		}

		time.Sleep(50 * time.Millisecond) // Don't chew CPU.
	}
}

func main() {
	log.Println("Starting TruthMachine v0.0.13")

	heartRate := make(chan int)
	go pulse(heartRate)

	addr := "localhost:8765"
	server := &osc.Server{Addr: addr}

	for _, endPoint := range []string{"/calibrate", "/interrogate", "/reset"} {
		server.Handle(endPoint, func(msg *osc.Message) {
			log.Println("Polygraph instruction: ")
			log.Println("http://10.0.1.3/arduino" + msg.Address)
			//_, err := http.Get("http://10.0.1.3/arduino" + msg.Address)
			_, err := http.Get("http://192.168.86.143/arduino" + msg.Address)

			if err != nil {
				log.Println("Unable to contact theatrical polygraph")
			}
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

		heartRate <- int(f)
		log.Println(fmt.Sprintf("/h (%.2f)", f))
	})

	log.Println("Creating Qlab endpoint: '/cue/gX/start'")
	http.HandleFunc("/g", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")

		f, err := strconv.ParseFloat(r.URL.Query()["v"][0], 32)
		if err != nil {
			log.Fatal("Unable to parse value argument for '/g'.")
		}

		b, err := strconv.ParseFloat(r.URL.Query()["b"][0], 32)
		if err != nil {
			log.Fatal("Unable to parse baseline argument for '/g'.")
		}

		baseline := 0.0
		highline := 1024.0

		if b > 0.1 {
			baseline = math.Max(0.0, b-(b*0.05))
			highline = b + (b * 0.2)
		}

		id := lerp(baseline, highline, f, 1, 20)
		client := osc.NewClient("localhost", 53000)
		msg := osc.NewMessage(fmt.Sprintf("/cue/g%d/start", id))
		client.Send(msg)
		log.Println(fmt.Sprintf("%s (%.2f)", msg.Address, f))
	})

	log.Println("Creating Qlab endpoint: '/cue/rX/start'")
	http.HandleFunc("/r", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")

		f, err := strconv.ParseFloat(r.URL.Query()["v"][0], 32)
		if err != nil {
			log.Fatal("Unable to parse argument for '/r'.")
		}

		id := lerp(0.0, 60.0, f, 1, 20)
		client := osc.NewClient("localhost", 53000)
		msg := osc.NewMessage(fmt.Sprintf("/cue/r%d/start", id))
		client.Send(msg)
		log.Println(fmt.Sprintf("%s (%.2f)", msg.Address, f))
	})

	log.Println("Creating Qlab endpoint: '/cue/lX/start'")
	http.HandleFunc("/l", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")

		f, err := strconv.ParseFloat(r.URL.Query()["v"][0], 32)
		if err != nil {
			log.Fatal("Unable to parse '/l' argument.")
		}

		id := lerp(0, 100, f, 1, 100)
		client := osc.NewClient("localhost", 53000)
		msg := osc.NewMessage(fmt.Sprintf("/cue/l%d/start", id))
		client.Send(msg)
		log.Println(fmt.Sprintf("%s (%.2f)", msg.Address, f))
	})

	log.Println("Creating calibrate notification")
	http.HandleFunc("/d", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
		log.Println("Calibration point recorded")
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
