// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flags

import "strconv"

// -- uint Value
type uintValue uint

func newUintValue(val uint, p *uint) *uintValue {
	*p = val
	return (*uintValue)(p)
}

func (i *uintValue) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, strconv.IntSize)
	if err != nil {
		err = numError(err)
	}
	*i = uintValue(v)
	return err
}

func (i *uintValue) Get() interface{} { return uint(*i) }

func (i *uintValue) String() string { return strconv.FormatUint(uint64(*i), 10) }

// UintVar defines a uint flag with specified name, default value, and usage string.
// The argument p points to a uint variable in which to store the value of the flag.
func (f *FlagSet) UintVar(p *uint, name string, alias rune, value uint, usage string, fn Callback) {
	f.Var(newUintValue(value, p), name, alias, usage, fn)
}

// UintVar defines a uint flag with specified name, default value, and usage string.
// The argument p points to a uint variable in which to store the value of the flag.
func UintVar(p *uint, name string, alias rune, value uint, usage string, fn Callback) {
	CommandLine.Var(newUintValue(value, p), name, alias, usage, fn)
}

// Uint defines a uint flag with specified name, default value, and usage string.
// The return value is the address of a uint variable that stores the value of the flag.
func (f *FlagSet) Uint(name string, alias rune, value uint, usage string, fn Callback) *uint {
	p := new(uint)
	f.UintVar(p, name, alias, value, usage, fn)
	return p
}

// Uint defines a uint flag with specified name, default value, and usage string.
// The return value is the address of a uint variable that stores the value of the flag.
func Uint(name string, alias rune, value uint, usage string, fn Callback) *uint {
	return CommandLine.Uint(name, alias, value, usage, fn)
}
