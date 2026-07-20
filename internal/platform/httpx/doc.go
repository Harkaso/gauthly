// Package httpx writes the HTTP responses shared by every handler of the
// service, so that status codes, headers and payload shapes stay identical
// across packages.
//
// It carries no business logic: callers supply the values, httpx encodes them
// and sets the headers every response is expected to carry.
package httpx
