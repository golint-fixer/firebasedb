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

package firebasedb

import (
	"fmt"
	"log"
	"sort"
)

func ExampleReference_Value() {

	const dinoFactsUrl = "https://dinosaur-facts.firebaseio.com/"

	type dinosaurFact struct {
		Appeared int64   `json:"appeared"`
		Height   float32 `json:"height"`
		Length   float32 `json:"length"`
		Order    string  `json:"order"`
		Vanished int64   `json:"vanished"`
		Weight   int32   `json:"weight"`
	}

	type dinosaurs map[string]dinosaurFact

	db := NewReference(dinoFactsUrl)
	if db.Error != nil {
		log.Fatalf("Error opening database: %v", db.Error)
	}
	var dinos = dinosaurs{}
	db.Ref("dinosaurs").Value(&dinos)
	var keys []string
	for k := range dinos {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("The %s's height is %v\n", k, dinos[k].Height)
	}
	// Output: The bruhathkayosaurus's height is 25
	// The lambeosaurus's height is 2.1
	// The linhenykus's height is 0.6
	// The pterodactyl's height is 0.6
	// The stegosaurus's height is 4
	// The triceratops's height is 3
}

func ExampleReference_StartAt() {
	const dinoFactsUrl = "https://dinosaur-facts.firebaseio.com/"

	type dinosaurFact struct {
		Appeared int64   `json:"appeared"`
		Height   float32 `json:"height"`
		Length   float32 `json:"length"`
		Order    string  `json:"order"`
		Vanished int64   `json:"vanished"`
		Weight   int32   `json:"weight"`
	}

	type dinosaurs map[string]dinosaurFact

	db := NewReference(dinoFactsUrl)
	if db.Error != nil {
		log.Fatalf("Error opening database: %v", db.Error)
	}
	var dinos = dinosaurs{}
	db.Ref("dinosaurs").OrderByChild("height").StartAt(3).EndAt(5).Value(&dinos)

	var keys []string
	for k := range dinos {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("The %s's height is %v\n", k, dinos[k].Height)
	}
	// Output: The stegosaurus's height is 4
	// The triceratops's height is 3
}

func ExampleReference_String() {
	const dinoFactsUrl = "https://dinosaur-facts.firebaseio.com/"
	db := NewReference(dinoFactsUrl)
	fmt.Println(db.Ref("dinosaurs").Child("stegosaurus").String())
	// Output: https://dinosaur-facts.firebaseio.com/dinosaurs/stegosaurus
}

func ExampleReference_PassKeepAlive() {
	const dinoFactsUrl = "https://dinosaur-facts.firebaseio.com/"
	db := NewReference(dinoFactsUrl)

	// Get an events subscription that filters out keep-alive events (default)
	s0, err := db.Ref("dinosaurs").PassKeepAlive(false).Subscribe()
	if err != nil {
		log.Fatal(err)
	}
	s0.Close()

	// Get an events subscription that includes keep-alive events
	s1, err := db.Ref("dinosaurs").PassKeepAlive(true).Subscribe()
	if err != nil {
		log.Fatal(err)
	}
	s1.Close()
}
