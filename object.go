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
Package AutoREST is a simplish automatic REST API generator for GORM.


*/
package autorest

import "errors"
import "reflect"
import "net/http"

import "gorm.io/gorm"

// RegisteredType represents a type that will have objects stored in the database and will be attached to a REST API.
type RegisteredType struct {
	Type reflect.Type
	DB   *gorm.DB
}

// RegisterType is a convenience method that fills out the fields of the RegisteredType and calls DB.AutoMigrate
// for you.
func RegisterType(t interface{}, DB *gorm.DB) *RegisteredType {
	rt := &RegisteredType{
		Type: reflect.TypeOf(t),
		DB:   DB,
	}
	DB.AutoMigrate(t)
	return rt
}

// A simple interface designed to match the decoder types provided by multiple standard library encoding packages.
// For example, json.Decoder satisfies this just fine.
type Decoder interface {
	// Decode *MUST* preserve existing data in the value it is given. Basically any keys not present in the
	// decoded data must be left set to their initial value. AFAIK all the standard library decoders do this.
	Decode(interface{}) error
}

// Logger is a very simple interface that exists solely so I don't have to tie the core to any given logging library.
type Logger interface {
	Println(v ...interface{})
}

// EndpointTypes is a simple bit field type used to specify which endpoints you want to mount from a given selection.
type EndpointTypes int

const (
	EndpointTypeCreate = EndpointTypes(1)  // POST
	EndpointTypeRead   = EndpointTypes(2)  // GET
	EndpointTypeList   = EndpointTypes(4)  // GET (with no ID specified)
	EndpointTypeUpdate = EndpointTypes(8)  // PUT
	EndpointTypeDelete = EndpointTypes(16) // DELETE
	EndpointTypeAll    = EndpointTypes(31) // All the end point types pre-combined, for convenience.
)

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// The "guts" of the system

// Create inserts the item into the database and returns a ready to return HTTP status code.
func (rt *RegisteredType) Create(log Logger, v Decoder) (code int) {
	newX := reflect.New(rt.Type).Interface()

	err := v.Decode(newX)
	if err != nil {
		// Could also be a 500, but unlikely. Most JSON errors are bad input data.
		log.Println(err)
		return http.StatusBadRequest
	}

	err = rt.DB.Create(newX).Error
	if err != nil {
		log.Println(err)

		return http.StatusInternalServerError
	}
	return http.StatusOK
}

// Read retrieves an item from the DB by ID, and returns it (or nil) along with a ready to use HTTP status code.
func (rt *RegisteredType) Read(log Logger, id uint64) (v interface{}, code int) {
	newX := reflect.New(rt.Type).Interface()

	err := rt.DB.First(newX, id).Error
	if err != nil {
		log.Println(err)

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, http.StatusNotFound
		}
		return nil, http.StatusInternalServerError
	}

	return newX, http.StatusOK
}

type listHeader struct {
	Page  int
	Limit int
	Total int64
	Data  interface{}
}

// List returns all the objects of a given type in the DB along with a ready to use HTTP status code.
func (rt *RegisteredType) List(log Logger, page, limit int) (v interface{}, code int) {
	newX := reflect.New(reflect.SliceOf(rt.Type)).Interface()
	newXS := reflect.New(reflect.SliceOf(rt.Type)).Interface()

	var count int64
	err := rt.DB.Model(newX).Count(&count).Error
	if err != nil {
		log.Println(err)

		return nil, http.StatusInternalServerError
	}

	db := rt.DB
	if page > 0 {
		db = db.Offset(page * limit)
	}
	if limit > 0 {
		db = db.Limit(limit)
	}
	err = db.Find(newXS).Error
	if err != nil {
		log.Println(err)

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, http.StatusNotFound
		}
		return nil, http.StatusInternalServerError
	}

	return &listHeader{page, limit, count, newXS}, http.StatusOK
}

// Update loads the specified value from the database, decodes the provided data over top of it, and then writes
// the result back out into the database.
func (rt *RegisteredType) Update(log Logger, id uint64, v Decoder) (code int) {
	newX := reflect.New(rt.Type).Interface()

	err := rt.DB.First(newX, id).Error
	if err != nil {
		log.Println(err)

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return http.StatusNotFound
		}
		return http.StatusInternalServerError
	}

	// Merge the encoded data into the item we just read.
	err = v.Decode(newX)
	if err != nil {
		// Could also be a 500, but unlikely. Most JSON errors are bad input data.
		log.Println(err)
		return http.StatusBadRequest
	}

	rt.DB.Save(newX)
	return http.StatusOK
}

// Delete removes the specified item from the database.
func (rt *RegisteredType) Delete(log Logger, id uint64) (code int) {
	newX := reflect.New(rt.Type).Interface()

	err := rt.DB.Delete(newX, id).Error
	if err != nil {
		log.Println(err)

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return http.StatusNotFound
		}
		return http.StatusInternalServerError
	}
	return http.StatusOK
}
