// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flags

import (
	"fmt"
	"os"
)

// failf prints to standard error a formatted error and usage message and
// returns the error.
func (f *FlagSet) failf(format string, a ...interface{}) error {
	err := fmt.Errorf(format, a...)
	if f.errorHandling != ContinueOnError {
		fmt.Fprintln(f.Output(), err)
	}
	// f.usage()
	return err
}

// usage calls the Usage method for the flag set if one is specified,
// or the appropriate default usage function otherwise.
func (f *FlagSet) usage() {
	if f.Usage == nil {
		f.defaultUsage()
	} else {
		f.Usage()
	}
}

func (f *FlagSet) cut() string {
	v := f.args[f.index]
	f.args = append(f.args[:f.index], f.args[f.index+1:]...)
	return v
}

func (f *FlagSet) setValue(flag *Flag, value string, hasValue bool) error {
	// boolean value is inverted unless a value is explicitly specified with "="
	if b, ok := flag.Value.(boolFlag); ok && b.IsBoolFlag() {
		if hasValue {
			if err := flag.Value.Set(value); err != nil {
				return f.failf("invalid boolean value %q for --%s: %v", value, flag.Name, err)
			}
		} else {
			if err := flag.Value.Set("true"); err != nil {
				return f.failf("invalid boolean flag %s: %v", flag.Name, err)
			}
		}
	} else if hasValue {
		if err := flag.Value.Set(value); err != nil {
			return f.failf("invalid boolean value %q for --%s: %v", value, flag.Name, err)
		}
	} else if f.index < len(f.args) {
		if err := flag.Value.Set(f.cut()); err != nil {
			return f.failf("invalid value %q for flag --%s: %v", value, flag.Name, err)
		}
	} else {
		return f.failf("flag needs an argument: --%s", flag.Name)
	}
	return nil
}

// parseOne parses one flag. It reports whether a flag was seen.
func (f *FlagSet) parseOne() (bool, error) {
	if len(f.args) == 0 {
		return false, nil
	}

	s := f.args[f.index]

	if len(s) > 1 && s[0] == '-' {
		f.cut()

		numMinuses := 1
		if s[1] == '-' {
			numMinuses++
			if len(s) == 2 { // "--" terminates the flags
				return false, nil
			}
		}

		name := s[numMinuses:]
		if len(name) == 0 || name[0] == '-' || name[0] == '=' {
			return false, f.failf("bad flag syntax: %s", s)
		}

		// it's a flag. does it have an argument?
		hasValue := false
		value := ""
		for i := 1; i < len(name); i++ { // equals cannot be first
			if name[i] == '=' {
				value = name[i+1:]
				hasValue = true
				name = name[0:i]
				break
			}
		}

		m := f.formal
		switch numMinuses {
		case 2:
			flag, alreadythere := m[name] // BUG
			if !alreadythere {
				if name == "help" { // special case for nice help message.
					f.usage()
					return false, ErrHelp
				}
				return false, f.failf("flag provided but not defined: --%s", name)
			}

			if err := f.setValue(flag, value, hasValue); err != nil {
				return false, err
			}

			if f.actual == nil {
				f.actual = make(map[string]*Flag)
			}
			f.actual[name] = flag

		case 1:
			for _, v := range name {
				longname, alreadythere := f.aliasToName[v]
				if !alreadythere {
					if v == 'h' { // special case for nice help message.
						f.usage()
						return false, ErrHelp
					}

					return false, f.failf("flag provided but not defined: -%c", v)
				}

				flag, alreadythere := m[longname]
				if alreadythere {
					if err := f.setValue(flag, value, hasValue); err != nil {
						return false, err
					}
				}

				if f.actual == nil {
					f.actual = make(map[string]*Flag)
				}
				f.actual[longname] = flag
			}
		}

	} else {
		if f.StopImmediate {
			return false, nil
		}
		f.index++
	}

	if f.index < len(f.args) {
		return true, nil
	}

	return false, nil
}

// Parse parses flag definitions from the argument list, which should not
// include the command name. Must be called after all flags in the FlagSet
// are defined and before flags are accessed by the program.
// The return value will be ErrHelp if -help or -h were set but not defined.
func (f *FlagSet) Parse(arguments []string) error {
	f.parsed = true
	f.index = 0
	f.args = arguments
	for {
		seen, err := f.parseOne()
		if seen {
			continue
		}
		if err == nil {
			break
		}
		switch f.errorHandling {
		case ContinueOnError:
			return err
		case ExitOnError:
			os.Exit(2)
		case PanicOnError:
			panic(err)
		}
	}
	return nil
}

// Parsed reports whether f.Parse has been called.
func (f *FlagSet) Parsed() bool {
	return f.parsed
}

// Parse parses the command-line flags from os.Args[1:]. Must be called
// after all flags are defined and before flags are accessed by the program.
func Parse() {
	// Ignore errors; CommandLine is set for ExitOnError.
	CommandLine.Parse(os.Args[1:])
}

// Parsed reports whether the command-line flags have been parsed.
func Parsed() bool {
	return CommandLine.Parsed()
}
