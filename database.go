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
	"net/http"
	"net/url"
	"path"
	"strings"
    "errors"
    "encoding/json"
)

type Reference struct { // TODO: check if we can replace by a simple URL
	url url.URL
}

// NewFirebaseDB opens a new Firebase Database connection using the URL u and the
// authentication auth. Currently, only the database secret can be used as auth.
func NewFirebaseDB(u, auth string) (Reference, error) {
	parsedUrl, err := url.Parse(u)
	if err != nil {
		return Reference{}, err
	} else {
		ref := Reference{
			url: *parsedUrl,
		}
		if auth != "" {
			ref = ref.withParam("auth", auth)
		}
		return ref, nil
	}
}

// jsonUrl is an internal function to build the URL for the REST API
// See https://firebase.google.com/docs/reference/rest/database/ "API Usage"
func (r Reference) jsonUrl() string {
	u := r.url
	u.Path = path.Clean(u.Path)
	if u.Path == "." {
		u.Path = "/.json"
	} else {
		u.Path = strings.Join([]string{u.Path, ".json"}, "")
	}
	return u.String()
}

// Values reads from the database and store the content in value. It gives an error
// if it the HTTP request fails or if it can't decode the JSON payload.
func (r Reference) Value(value interface{}) error {
	response, err := http.Get(r.jsonUrl())
	if err != nil {
		return err
	}
    defer response.Body.Close()
    if response.StatusCode != 200 {
        return errors.New(response.Status)
    }
    d := json.NewDecoder(response.Body)
    return d.Decode(value)
 }
