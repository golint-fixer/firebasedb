// Copyright 2016 Jacques Supcik
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// This code implements the techniques explained in "Advanced Go
// Concurrency Patterns" from Sameer Ajmani.
// https://blog.golang.org/advanced-go-concurrency-patterns

package firebasedb

import (
	"bufio"
	"errors"
	"io"
	"net/http"
	"strings"
	"encoding/json"
)

type Event struct {
	Type string
	data string
}

func (e Event) Value(v interface{}) (path string, err error) {
	var p struct {
		Path string      `json:"path"`
		Data interface{} `json:"data"`
	}
	p.Data = &v
	err = json.Unmarshal([]byte(e.data), &p)
	if err != nil {
		path = ""
	} else {
		path = p.Path
	}
	return path, err
}

type Subscription interface {
	Events() <-chan Event // stream of Events
	Close() error         // shuts down the stream
}

// sub implements the Subscription interface.
type sub struct {
	reader  io.ReadCloser   // from the HTTP request's body
	events  chan Event      // sends events to the user
	closing chan chan error // for Close
}

func (r Reference) Subscribe() (Subscription, error) {
	req, err := http.NewRequest("GET", r.jsonUrl(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "text/event-stream")
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		response.Body.Close()
		return nil, errors.New(response.Status)
	}
	s := &sub{
		reader:  response.Body,
		events:  make(chan Event),      // for Events
		closing: make(chan chan error), // for Close
	}
	go s.loop()
	return s, nil
}

func (s *sub) Events() <-chan Event {
	return s.events
}

func (s *sub) Close() error {
	errChan := make(chan error)
	s.closing <- errChan
	return <-errChan
}

func (s *sub) loop() {
	type fetchData struct {
		event Event
		err   error
	}

	var fetchEvent chan fetchData
	var pending []Event
	var err error

	go func() { // read the payload and feed the fetchEvent channel
		payload := make([]string, 2)
		lineCount := 0
		r := bufio.NewReader(s.reader)
		for {
			line, err := r.ReadString('\n')
			if err != nil {
				break
			}
			line = strings.Trim(line, " \r\n")
			if len(line) == 0 {
				// empty line
				if lineCount == len(payload) {
					if !strings.HasPrefix(payload[0], "event:") {
						fetchEvent <- fetchData{
							event: Event{},
							err:   errors.New("First line does not start with event:"),
						}
					} else if !strings.HasPrefix(payload[0], "data:") {
						fetchEvent <- fetchData{
							event: Event{},
							err:   errors.New("Second line does not start with data:"),
						}
					} else {
						fetchEvent <- fetchData{
							event: Event{
								Type: strings.Trim(strings.TrimPrefix(payload[0], "event:"), " \r\n"),
								data: strings.Trim(strings.TrimPrefix(payload[1], "data:"), " \r\n"),
							},
							err: nil,
						}
					}
				} else {
					fetchEvent <- fetchData{event: Event{}, err: errors.New("Bad formated body")}
				}
				lineCount = 0
			} else { // line is not empty
				if lineCount < len(payload) {
					payload[lineCount] = line
					lineCount++
				}
			}
		}
	}()

	for {
		var first Event
		var events chan Event
		if len(pending) > 0 {
			first = pending[0]
			events = s.events // enable send case
		}

		select {
		case eventc := <-fetchEvent:
			if eventc.err != nil {
				// break
			} else {
				pending = append(pending, eventc.event)
			}
		case errc := <-s.closing:
			errc <- err
			s.reader.Close()
			close(s.events)
			return
		case events <- first:
			pending = pending[1:]
		}
	}
}