// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flags

// -- string Value
type stringValue string

func newStringValue(val string, p *string) *stringValue {
	*p = val
	return (*stringValue)(p)
}

func (s *stringValue) Set(val string) error {
	*s = stringValue(val)
	return nil
}

func (s *stringValue) Get() interface{} { return string(*s) }

func (s *stringValue) String() string { return string(*s) }

// StringVar defines a string flag with specified name, default value, and usage string.
// The argument p points to a string variable in which to store the value of the flag.
func (f *FlagSet) StringVar(p *string, name string, alias rune, value string, usage string, fn Callback) {
	f.Var(newStringValue(value, p), name, alias, usage, fn)
}

// StringVar defines a string flag with specified name, default value, and usage string.
// The argument p points to a string variable in which to store the value of the flag.
func StringVar(p *string, name string, alias rune, value string, usage string, fn Callback) {
	CommandLine.Var(newStringValue(value, p), name, alias, usage, fn)
}

// String defines a string flag with specified name, default value, and usage string.
// The return value is the address of a string variable that stores the value of the flag.
func (f *FlagSet) String(name string, alias rune, value string, usage string, fn Callback) *string {
	p := new(string)
	f.StringVar(p, name, alias, value, usage, fn)
	return p
}

// String defines a string flag with specified name, default value, and usage string.
// The return value is the address of a string variable that stores the value of the flag.
func String(name string, alias rune, value string, usage string, fn Callback) *string {
	return CommandLine.String(name, alias, value, usage, fn)
}
