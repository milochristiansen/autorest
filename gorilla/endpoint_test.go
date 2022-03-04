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

package gorilla_test

import "testing"

import "strings"
import "net/http"
import "net/http/httptest"

import "github.com/milochristiansen/autorest"
import "github.com/milochristiansen/autorest/gorilla"

import "github.com/gorilla/mux"

import "github.com/milochristiansen/sessionlogger"

import "github.com/glebarez/sqlite"
import "gorm.io/gorm"

// TestBasicFunction... Tests the basic functionality. No detailed testing is done here,
// just a quick top level sanity check.
func TestBasicFunction(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	router := mux.NewRouter()

	rt := autorest.RegisterType(&TestType{}, db)
	gorilla.CreateEndpoints(rt, autorest.EndpointTypeAll, "/test", router, &sessionlogger.Config{})

	// Create
	r, err := http.NewRequest("POST", "/test", strings.NewReader(`{"String": "test", "Int": 5}`))
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("Status not OK: %v", http.StatusText(w.Code))
	}

	// List (all)
	r, err = http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	w = httptest.NewRecorder()

	router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("Status not OK: %v", http.StatusText(w.Code))
	}
	if w.Body.String() != `{"Page":0,"Limit":0,"Total":1,"Data":[{"ID":1,"String":"test","Int":5}]}`+"\n" {
		t.Fatalf("Body value is unexpected: %#v", w.Body.String())
	}

	// Read
	r, err = http.NewRequest("GET", "/test/1", nil)
	if err != nil {
		t.Fatal(err)
	}
	w = httptest.NewRecorder()

	router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("Status not OK: %v", http.StatusText(w.Code))
	}
	if w.Body.String() != `{"ID":1,"String":"test","Int":5}`+"\n" {
		t.Fatalf("Body value is unexpected: %#v", w.Body.String())
	}

	// Update
	r, err = http.NewRequest("PUT", "/test/1", strings.NewReader(`{"Int": 10}`))
	if err != nil {
		t.Fatal(err)
	}
	w = httptest.NewRecorder()

	router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("Status not OK: %v", http.StatusText(w.Code))
	}

	// Read again to confirm
	r, err = http.NewRequest("GET", "/test/1", nil)
	if err != nil {
		t.Fatal(err)
	}
	w = httptest.NewRecorder()

	router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("Status not OK: %v", http.StatusText(w.Code))
	}
	if w.Body.String() != `{"ID":1,"String":"test","Int":10}`+"\n" {
		t.Fatalf("Body value is unexpected: %#v", w.Body.String())
	}

	// Delete
	r, err = http.NewRequest("DELETE", "/test/1", nil)
	if err != nil {
		t.Fatal(err)
	}
	w = httptest.NewRecorder()

	router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("Status not OK: %v", http.StatusText(w.Code))
	}

	// Read again to confirm
	r, err = http.NewRequest("GET", "/test/1", nil)
	if err != nil {
		t.Fatal(err)
	}
	w = httptest.NewRecorder()

	router.ServeHTTP(w, r)
	if w.Code != http.StatusNotFound {
		t.Fatalf("Status not OK: %v", http.StatusText(w.Code))
	}
}

type TestType struct {
	ID     uint
	String string
	Int    int
}
