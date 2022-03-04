/*
Copyright 2022 by Milo Christiansen

This software is provided 'as-is', without any express or implied warranty. In
no event will the authors be held liable for any damages arising from the use of
this software.

Permission is granted to anyone to use this software for any purpose, including
commercial applications, and to alter it and redistribute it freely, subject to
the following restrictions:

1. The origin of this software must not be misrepresented; you must not claim
that you wrote the original software. If you use this software in a product, an
acknowledgment in the product documentation would be appreciated but is not
required.

2. Altered source versions must be plainly marked as such, and must not be
misrepresented as being the original software.

3. This notice may not be removed or altered from any source distribution.
*/

/*
This package is an implementation of AutoREST that uses net/http, encoding/json, and my own logger wrapper thingy.
*/
package vanilla

import "strconv"
import "net/http"
import "encoding/json"

import "github.com/milochristiansen/sessionlogger"
import "github.com/milochristiansen/autorest"

// CreateEndpoints mounts the requested endpoint types at the given path on the given Router.
// All the endpoints go on the same path, but with different methods. The path is a prefix, and should
// not have a trailing slash.
//
// Create: POST the object
// Read: GET with ?id=<id>
// List: GET without an ID (?page=x&limit=y optional)
// Update: PUT the object (may be partial) with ?id=<id>
// Delete: DELETE with ?id=<id>
//
// I was too lazy to put the IDs on the path like in the gorilla/mux version, so this version has to make do with variables.
// The gorilla/mux version is nicer, use that.
func CreateEndpoints(rt *autorest.RegisteredType, desiredEndpoints autorest.EndpointTypes, path string, router *http.ServeMux, logc *sessionlogger.Config) {
	router.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			// POST /
			// Payload: Full object, JSON encoded
			// Returns: Nothing.
			if desiredEndpoints&autorest.EndpointTypeCreate != 0 {
				log := logc.NewSessionLogger("POST:" + path)

				w.WriteHeader(rt.Create(log.E, json.NewDecoder(r.Body)))
			}
			return
		case "GET":
			idS := r.FormValue("id")
			if idS == "" {
				// GET /
				// GET /?page=x&limit=y
				// Returns: {Page: 0, Limit: 0, Total: 0, Data: {}}
				if desiredEndpoints&autorest.EndpointTypeList != 0 {
					log := logc.NewSessionLogger("GET:" + path)

					var err error
					page, limit := 0, 0

					if pageS := r.FormValue("page"); pageS != "" {
						page, err = strconv.Atoi(pageS)
						if err != nil {
							log.E.Println(err)

							w.WriteHeader(http.StatusBadRequest)
							return
						}
					}

					if limitS := r.FormValue("limit"); limitS != "" {
						limit, err = strconv.Atoi(limitS)
						if err != nil {
							log.E.Println(err)

							w.WriteHeader(http.StatusBadRequest)
							return
						}
					}

					v, code := rt.List(log.E, page, limit)

					w.WriteHeader(code)
					if v != nil {
						err := json.NewEncoder(w).Encode(v)
						if err != nil {
							log.E.Println(err)
						}
					}
				}
				return
			}

			// GET /?id=<id>
			// Returns: Full object with given id
			if desiredEndpoints&autorest.EndpointTypeRead != 0 {
				log := logc.NewSessionLogger("GET:" + path + "?id=<id>")

				id, err := strconv.ParseUint(idS, 10, 0)
				if err != nil {
					log.E.Println(err)

					w.WriteHeader(http.StatusBadRequest)
					return
				}

				v, code := rt.Read(log.E, id)
				w.WriteHeader(code)
				if v != nil {
					err := json.NewEncoder(w).Encode(v)
					if err != nil {
						log.E.Println(err)
					}
				}
			}
		case "PUT":
			// PUT /?id=<id>
			// Payload: Partial JSON object
			// Returns: Nothing.
			if desiredEndpoints&autorest.EndpointTypeUpdate != 0 {
				log := logc.NewSessionLogger("PUT:" + path + "?id=<id>")

				id, err := strconv.ParseUint(r.FormValue("id"), 10, 0)
				if err != nil {
					log.E.Println(err)

					w.WriteHeader(http.StatusBadRequest)
					return
				}

				w.WriteHeader(rt.Update(log.E, id, json.NewDecoder(r.Body)))
			}
		case "DELETE":
			// DELETE /?id=<id>
			// Returns: Nothing.
			if desiredEndpoints&autorest.EndpointTypeDelete != 0 {
				log := logc.NewSessionLogger("DELETE:" + path + "?id=<id>")

				id, err := strconv.ParseUint(r.FormValue("id"), 10, 0)
				if err != nil {
					log.E.Println(err)

					w.WriteHeader(http.StatusBadRequest)
					return
				}

				w.WriteHeader(rt.Delete(log.E, id))
			}
		}
	})

}
