// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package safehttp

import (
	"errors"
	"net/http"
	"net/textproto"
)

// Header represents the key-value pairs in an HTTP header.
// The keys will be in canonical form, as returned by
// textproto.CanonicalMIMEHeaderKey.
type Header struct {
	wrapped   http.Header
	immutable map[string]bool
}

func newHeader(h http.Header) Header {
	return Header{wrapped: h, immutable: map[string]bool{}}
}

// MarkImmutable marks the header with the given name as immutable.
// The name is first canonicalized using textproto.CanonicalMIMEHeaderKey.
// This header is now read-only. If the header previously was
// undefined, then it will forever be undefined. Other methods in
// the struct can't write to, change or delete the header with this
// name. These methods will instead fail when applied on an immutable
// header.
func (h Header) MarkImmutable(name string) {
	name = textproto.CanonicalMIMEHeaderKey(name)
	h.immutable[name] = true
}

// Set sets the header with the given name to the given value.
// The name is first canonicalized using textproto.CanonicalMIMEHeaderKey.
// This method first removes all other values associated with this
// header before setting the new value. Returns an error when
// applied on immutable headers or on the Set-Cookie header.
func (h Header) Set(name, value string) error {
	name = textproto.CanonicalMIMEHeaderKey(name)
	if err := h.writableHeader(name); err != nil {
		return err
	}
	h.wrapped.Set(name, value)
	return nil
}

// Add adds a new header with the given name and the given value to
// the collection of headers. The name is first canonicalized using
// textproto.CanonicalMIMEHeaderKey. Returns an error when applied
// on immutable headers or on the Set-Cookie header.
func (h Header) Add(name, value string) error {
	name = textproto.CanonicalMIMEHeaderKey(name)
	if err := h.writableHeader(name); err != nil {
		return err
	}
	h.wrapped.Add(name, value)
	return nil
}

// Del deletes all headers with the given name. The name is first
// canonicalized using textproto.CanonicalMIMEHeaderKey. Returns an
// error when applied on immutable headers or on the Set-Cookie
// header.
func (h Header) Del(name string) error {
	name = textproto.CanonicalMIMEHeaderKey(name)
	if err := h.writableHeader(name); err != nil {
		return err
	}
	h.wrapped.Del(name)
	return nil
}

// Get returns the value of the first header with the given name.
// The name is first canonicalized using textproto.CanonicalMIMEHeaderKey.
// If no header exists with the given name then "" is returned.
func (h Header) Get(name string) string {
	return h.wrapped.Get(name)
}

// Values returns all the values of all the headers with the given name.
// The name is first canonicalized using textproto.CanonicalMIMEHeaderKey.
// The values are returned in the same order as they were sent in the request.
// The values are returned as a copy of the original slice of strings in
// the internal header map. This is to prevent modification of the original
// slice. If no header exists with the given name then an empty slice is
// returned.
func (h Header) Values(name string) []string {
	v := h.wrapped.Values(name)
	clone := make([]string, len(v))
	copy(clone, v)
	return clone
}

// SetCookie adds the cookie provided as a Set-Cookie header in the header
// collection. If the cookie is nil or cookie.Name is invalid, no header is
// added. This is the only method that can modify the Set-Cookie header.
// If other methods try to modify the header they will return errors.
// TODO: Replace http.Cookie with safehttp.Cookie.
func (h Header) SetCookie(cookie *http.Cookie) {
	if v := cookie.String(); v != "" {
		h.wrapped.Add("Set-Cookie", v)
	}
}

// TODO: Add Write, WriteSubset and Clone when needed.

// writableHeader assumes that the given name already has been canonicalized
// using textproto.CanonicalMIMEHeaderKey.
func (h Header) writableHeader(name string) error {
	// TODO(@mattiasgrenfeldt, @kele, @empijei): Think about how this should
	// work during legacy conversions.
	if name == "Set-Cookie" {
		return errors.New("can't write to Set-Cookie header")
	}
	if h.immutable[name] {
		return errors.New("immutable header")
	}
	return nil
}
