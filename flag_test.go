// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flags_test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	. "github.com/saihon/flags"
)

func boolString(s string) string {
	if s == "0" {
		return "false"
	}
	return "true"
}

func TestEverything(t *testing.T) {
	ResetForTesting(nil)
	Bool("test_bool", 0, false, "bool value", nil)
	Int("test_int", 0, 0, "int value", nil)
	Int64("test_int64", 0, 0, "int64 value", nil)
	Uint("test_uint", 0, 0, "uint value", nil)
	Uint64("test_uint64", 0, 0, "uint64 value", nil)
	String("test_string", 0, "0", "string value", nil)
	Float64("test_float64", 0, 0, "float64 value", nil)
	Duration("test_duration", 0, 0, "time.Duration value", nil)

	m := make(map[string]*Flag)
	desired := "0"
	visitor := func(f *Flag) {
		if len(f.Name) > 5 && f.Name[0:5] == "test_" {
			m[f.Name] = f
			ok := false
			switch {
			case f.Value.String() == desired:
				ok = true
			case f.Name == "test_bool" && f.Value.String() == boolString(desired):
				ok = true
			case f.Name == "test_duration" && f.Value.String() == desired+"s":
				ok = true
			}
			if !ok {
				t.Error("Visit: bad value", f.Value.String(), "for", f.Name)
			}
		}
	}
	VisitAll(visitor)
	if len(m) != 8 {
		t.Error("VisitAll misses some flags")
		for k, v := range m {
			t.Log(k, *v)
		}
	}
	m = make(map[string]*Flag)
	Visit(visitor)
	if len(m) != 0 {
		t.Errorf("Visit sees unset flags")
		for k, v := range m {
			t.Log(k, *v)
		}
	}
	// Now set all flags
	Set("test_bool", "true")
	Set("test_int", "1")
	Set("test_int64", "1")
	Set("test_uint", "1")
	Set("test_uint64", "1")
	Set("test_string", "1")
	Set("test_float64", "1")
	Set("test_duration", "1s")
	desired = "1"
	Visit(visitor)
	if len(m) != 8 {
		t.Error("Visit fails after set")
		for k, v := range m {
			t.Log(k, *v)
		}
	}
	// Now test they're visited in sort order.
	var flagNames []string
	Visit(func(f *Flag) { flagNames = append(flagNames, f.Name) })
	if !sort.StringsAreSorted(flagNames) {
		t.Errorf("flag names not sorted: %v", flagNames)
	}
}

func TestGet(t *testing.T) {
	ResetForTesting(nil)
	Bool("test_bool", 0, true, "bool value", nil)
	Int("test_int", 0, 1, "int value", nil)
	Int64("test_int64", 0, 2, "int64 value", nil)
	Uint("test_uint", 0, 3, "uint value", nil)
	Uint64("test_uint64", 0, 4, "uint64 value", nil)
	String("test_string", 0, "5", "string value", nil)
	Float64("test_float64", 0, 6, "float64 value", nil)
	Duration("test_duration", 0, 7, "time.Duration value", nil)

	visitor := func(f *Flag) {
		if len(f.Name) > 5 && f.Name[0:5] == "test_" {
			g, ok := f.Value.(Getter)
			if !ok {
				t.Errorf("Visit: value does not satisfy Getter: %T", f.Value)
				return
			}
			switch f.Name {
			case "test_bool":
				ok = g.Get() == true
			case "test_int":
				ok = g.Get() == int(1)
			case "test_int64":
				ok = g.Get() == int64(2)
			case "test_uint":
				ok = g.Get() == uint(3)
			case "test_uint64":
				ok = g.Get() == uint64(4)
			case "test_string":
				ok = g.Get() == "5"
			case "test_float64":
				ok = g.Get() == float64(6)
			case "test_duration":
				ok = g.Get() == time.Duration(7)
			}
			if !ok {
				t.Errorf("Visit: bad value %T(%v) for %s", g.Get(), g.Get(), f.Name)
			}
		}
	}
	VisitAll(visitor)
}

func TestUsage(t *testing.T) {
	// called := false
	// ResetForTesting(func() { called = true })
	ResetForTesting(nil)
	if CommandLine.Parse([]string{"-x"}) == nil {
		t.Error("parse did not fail for unknown flag")
	}
	// if !called {
	// 	t.Error("did not call Usage for unknown flag")
	// }
}

func testParse(f *FlagSet, t *testing.T) {
	if f.Parsed() {
		t.Error("f.Parse() = true before Parse")
	}
	boolFlag := f.Bool("bool", 0, false, "bool value", nil)
	bool2Flag := f.Bool("bool2", 0, false, "bool2 value", nil)
	intFlag := f.Int("int", 0, 0, "int value", nil)
	int64Flag := f.Int64("int64", 0, 0, "int64 value", nil)
	uintFlag := f.Uint("uint", 0, 0, "uint value", nil)
	uint64Flag := f.Uint64("uint64", 0, 0, "uint64 value", nil)
	stringFlag := f.String("string", 0, "0", "string value", nil)
	float64Flag := f.Float64("float64", 0, 0, "float64 value", nil)
	durationFlag := f.Duration("duration", 0, 5*time.Second, "time.Duration value", nil)
	extra := "one-extra-argument"
	args := []string{
		"--bool",
		"--bool2=true",
		"--int", "22",
		"--int64", "0x23",
		"--uint", "24",
		"--uint64", "25",
		"--string", "hello",
		"--float64", "2718e28",
		"--duration", "2m",
		extra,
	}
	if err := f.Parse(args); err != nil {
		t.Fatal(err)
	}
	if !f.Parsed() {
		t.Error("f.Parse() = false after Parse")
	}
	if *boolFlag != true {
		t.Error("bool flag should be true, is ", *boolFlag)
	}
	if *bool2Flag != true {
		t.Error("bool2 flag should be true, is ", *bool2Flag)
	}
	if *intFlag != 22 {
		t.Error("int flag should be 22, is ", *intFlag)
	}
	if *int64Flag != 0x23 {
		t.Error("int64 flag should be 0x23, is ", *int64Flag)
	}
	if *uintFlag != 24 {
		t.Error("uint flag should be 24, is ", *uintFlag)
	}
	if *uint64Flag != 25 {
		t.Error("uint64 flag should be 25, is ", *uint64Flag)
	}
	if *stringFlag != "hello" {
		t.Error("string flag should be `hello`, is ", *stringFlag)
	}
	if *float64Flag != 2718e28 {
		t.Error("float64 flag should be 2718e28, is ", *float64Flag)
	}
	if *durationFlag != 2*time.Minute {
		t.Error("duration flag should be 2m, is ", *durationFlag)
	}
	if len(f.Args()) != 1 {
		t.Error("expected one argument, got", len(f.Args()))
	} else if f.Args()[0] != extra {
		t.Errorf("expected argument %q got %q", extra, f.Args()[0])
	}
}

func TestParse(t *testing.T) {
	ResetForTesting(func() { t.Error("bad parse") })
	testParse(CommandLine, t)
}

func TestFlagSetParse(t *testing.T) {
	testParse(NewFlagSet("test", ContinueOnError, true), t)
}

// Declare a user-defined flag type.
type flagVar []string

func (f *flagVar) String() string {
	return fmt.Sprint([]string(*f))
}

func (f *flagVar) Set(value string) error {
	*f = append(*f, value)
	return nil
}

func TestUserDefined(t *testing.T) {
	var flags FlagSet
	flags.Init("test", ContinueOnError, true)
	var v flagVar
	flags.Var(&v, "v", 'v', "usage", nil)
	if err := flags.Parse([]string{"--v", "1", "--v=2", "-v", "3", "-v=4"}); err != nil {
		t.Error(err)
	}
	if len(v) != 4 {
		t.Fatal("expected 3 args; got ", len(v))
	}
	expect := "[1 2 3 4]"
	if v.String() != expect {
		t.Errorf("expected value %q got %q", expect, v.String())
	}
}

func TestUserDefinedForCommandLine(t *testing.T) {
	const help = "HELP"
	var result string
	ResetForTesting(func() { result = help })
	Usage()
	if result != help {
		t.Fatalf("got %q; expected %q", result, help)
	}
}

// Declare a user-defined boolean flag type.
type boolFlagVar struct {
	count int
}

func (b *boolFlagVar) String() string {
	return fmt.Sprintf("%d", b.count)
}

func (b *boolFlagVar) Set(value string) error {
	if value == "true" {
		b.count++
	}
	return nil
}

func (b *boolFlagVar) IsBoolFlag() bool {
	return b.count < 4
}

func TestUserDefinedBool(t *testing.T) {
	var flags FlagSet
	flags.Init("test", ContinueOnError, true)
	var b boolFlagVar
	var err error
	flags.Var(&b, "bool", 'b', "usage", nil)
	if err = flags.Parse([]string{"--bool", "-b", "-b", "--bool=true", "-b=false", "--bool", "barg", "-b"}); err != nil {
		if b.count < 4 {
			t.Error(err)
		}
	}

	if b.count != 4 {
		t.Errorf("want: %d; got: %d", 4, b.count)
	}

	if err == nil {
		t.Error("expected error; got none")
	}
}

func TestSetOutput(t *testing.T) {
	var flags FlagSet
	var buf bytes.Buffer
	flags.SetOutput(&buf)
	flags.Init("test", ContinueOnError, true)
	flags.Parse([]string{"--unknown"})
	if out := buf.String(); !strings.Contains(out, "--unknown") {
		t.Logf("expected output mentioning unknown; got %q", out)
	}
}

// This tests that one can reset the flags. This still works but not well, and is
// superseded by FlagSet.
func TestChangingArgs(t *testing.T) {
	ResetForTesting(func() { t.Fatal("bad parse") })
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", "--before", "subcmd", "--after", "args"}
	before := Bool("before", 0, false, "", nil)
	if err := CommandLine.Parse(os.Args[1:]); err != nil {
		t.Fatal(err)
	}
	cmd := Arg(0)
	os.Args = Args()
	after := Bool("after", 0, false, "", nil)
	Parse()
	args := Args()

	if !*before || cmd != "subcmd" || !*after || len(args) != 1 || args[0] != "args" {
		t.Fatalf("expected true subcmd true [args] got %v %v %v %v", *before, cmd, *after, args)
	}
}

// Test that -help invokes the usage message and returns ErrHelp.
func TestHelp(t *testing.T) {
	var helpCalled = false
	fs := NewFlagSet("help test", ContinueOnError, true)
	fs.Usage = func() { helpCalled = true }
	var flag bool
	fs.BoolVar(&flag, "flag", 0, false, "regular flag", nil)
	// Regular flag invocation should work
	err := fs.Parse([]string{"--flag=true"})
	if err != nil {
		t.Fatal("expected no error; got ", err)
	}
	if !flag {
		t.Error("flag was not set by -flag")
	}
	if helpCalled {
		t.Error("help called for regular flag")
		helpCalled = false // reset for next test
	}
	// Help flag should work as expected.
	err = fs.Parse([]string{"--help"})
	if err == nil {
		t.Fatal("error expected")
	}
	if err != ErrHelp {
		t.Fatal("expected ErrHelp; got ", err)
	}
	if !helpCalled {
		t.Fatal("help was not called")
	}
	// If we define a help flag, that should override.
	var help bool
	fs.BoolVar(&help, "help", 0, false, "help flag", nil)
	helpCalled = false
	err = fs.Parse([]string{"--help"})
	if err != nil {
		t.Fatal("expected no error for defined -help; got ", err)
	}
	if helpCalled {
		t.Fatal("help was called; should not have been for defined help flag")
	}
}

// const defaultOutput = `  -A	for bootstrapping, allow 'any' type
//   -Alongflagname
//     	disable bounds checking
//   -C	a boolean defaulting to true (default true)
//   -D path
//     	set relative path for local imports
//   -E string
//     	issue 23543 (default "0")
//   -F number
//     	a non-zero number (default 2.7)
//   -G float
//     	a float that defaults to zero
//   -M string
//     	a multiline
//     	help
//     	string
//   -N int
//     	a non-zero int (default 27)
//   -O	a flag
//     	multiline help string (default true)
//   -Z int
//     	an int that defaults to zero
//   -maxT timeout
//     	set timeout for dial
// `

// func TestPrintDefaults(t *testing.T) {
// 	fs := NewFlagSet("print defaults test", ContinueOnError)
// 	var buf bytes.Buffer
// 	fs.SetOutput(&buf)
// 	fs.Bool("A", 0, false, "for bootstrapping, allow 'any' type")
// 	fs.Bool("Alongflagname", 0, false, "disable bounds checking")
// 	fs.Bool("C", 0, true, "a boolean defaulting to true")
// 	fs.String("D", 0, "", "set relative `path` for local imports")
// 	fs.String("E", 0, "0", "issue 23543")
// 	fs.Float64("F", 0, 2.7, "a non-zero `number`")
// 	fs.Float64("G", 0, 0, "a float that defaults to zero")
// 	fs.String("M", 0, "", "a multiline\nhelp\nstring")
// 	fs.Int("N", 0, 27, "a non-zero int")
// 	fs.Bool("O", 0, true, "a flag\nmultiline help string")
// 	fs.Int("Z", 0, 0, "an int that defaults to zero")
// 	fs.Duration("maxT", 0, 0, "set `timeout` for dial")
// 	fs.PrintDefaults()
// 	got := buf.String()
// 	if got != defaultOutput {
// 		t.Errorf("got %q want %q\n", got, defaultOutput)
// 	}
// }

// Issue 19230: validate range of Int and Uint flag values.
func TestIntFlagOverflow(t *testing.T) {
	if strconv.IntSize != 32 {
		return
	}
	ResetForTesting(nil)
	Int("i", 0, 0, "", nil)
	Uint("u", 0, 0, "", nil)
	if err := Set("i", "2147483648"); err == nil {
		t.Error("unexpected success setting Int")
	}
	if err := Set("u", "4294967296"); err == nil {
		t.Error("unexpected success setting Uint")
	}
}

// // Issue 20998: Usage should respect CommandLine.output.
// func TestUsageOutput(t *testing.T) {
// 	ResetForTesting(DefaultUsage)
// 	var buf bytes.Buffer
// 	CommandLine.SetOutput(&buf)
// 	defer func(old []string) { os.Args = old }(os.Args)
// 	os.Args = []string{"app", "-i=1", "-unknown"}
// 	Parse()
// 	const want = "flag provided but not defined: -i\nUsage of app:\n"
// 	if got := buf.String(); got != want {
// 		t.Errorf("output = %q; want %q", got, want)
// 	}
// }

func TestGetters(t *testing.T) {
	expectedName := "flag set"
	expectedErrorHandling := ContinueOnError
	expectedOutput := io.Writer(os.Stderr)
	fs := NewFlagSet(expectedName, expectedErrorHandling, true)

	if fs.Name() != expectedName {
		t.Errorf("unexpected name: got %s, expected %s", fs.Name(), expectedName)
	}
	if fs.ErrorHandling() != expectedErrorHandling {
		t.Errorf("unexpected ErrorHandling: got %d, expected %d", fs.ErrorHandling(), expectedErrorHandling)
	}
	if fs.Output() != expectedOutput {
		t.Errorf("unexpected output: got %#v, expected %#v", fs.Output(), expectedOutput)
	}

	expectedName = "gopher"
	expectedErrorHandling = ExitOnError
	expectedOutput = os.Stdout
	fs.Init(expectedName, expectedErrorHandling, true)
	fs.SetOutput(expectedOutput)

	if fs.Name() != expectedName {
		t.Errorf("unexpected name: got %s, expected %s", fs.Name(), expectedName)
	}
	if fs.ErrorHandling() != expectedErrorHandling {
		t.Errorf("unexpected ErrorHandling: got %d, expected %d", fs.ErrorHandling(), expectedErrorHandling)
	}
	if fs.Output() != expectedOutput {
		t.Errorf("unexpected output: got %v, expected %v", fs.Output(), expectedOutput)
	}
}

func TestParseError(t *testing.T) {
	for _, typ := range []string{"bool", "int", "int64", "uint", "uint64", "float64", "duration"} {
		fs := NewFlagSet("parse error test", ContinueOnError, true)
		fs.SetOutput(ioutil.Discard)
		_ = fs.Bool("bool", 0, false, "", nil)
		_ = fs.Int("int", 0, 0, "", nil)
		_ = fs.Int64("int64", 0, 0, "", nil)
		_ = fs.Uint("uint", 0, 0, "", nil)
		_ = fs.Uint64("uint64", 0, 0, "", nil)
		_ = fs.Float64("float64", 0, 0, "", nil)
		_ = fs.Duration("duration", 0, 0, "", nil)
		// Strings cannot give errors.
		args := []string{"--" + typ + "=x"}
		err := fs.Parse(args) // x is not a valid setting for any flag.
		if err == nil {
			t.Errorf("Parse(%q)=%v; expected parse error", args, err)
			continue
		}
		if !strings.Contains(err.Error(), "invalid") || !strings.Contains(err.Error(), "parse error") {
			t.Errorf("Parse(%q)=%v; expected parse error", args, err)
		}
	}
}

func TestRangeError(t *testing.T) {
	bad := []string{
		"--int=123456789012345678901",
		"--int64=123456789012345678901",
		"--uint=123456789012345678901",
		"--uint64=123456789012345678901",
		"--float64=1e1000",
	}
	for _, arg := range bad {
		fs := NewFlagSet("parse error test", ContinueOnError, true)
		fs.SetOutput(ioutil.Discard)
		_ = fs.Int("int", 0, 0, "", nil)
		_ = fs.Int64("int64", 0, 0, "", nil)
		_ = fs.Uint("uint", 0, 0, "", nil)
		_ = fs.Uint64("uint64", 0, 0, "", nil)
		_ = fs.Float64("float64", 0, 0, "", nil)
		// Strings cannot give errors, and bools and durations do not return strconv.NumError.
		err := fs.Parse([]string{arg})
		if err == nil {
			t.Errorf("Parse(%q)=%v; expected range error", arg, err)
			continue
		}
		if !strings.Contains(err.Error(), "invalid") || !strings.Contains(err.Error(), "value out of range") {
			t.Errorf("Parse(%q)=%v; expected range error", arg, err)
		}
	}
}

func TestShortFlag(t *testing.T) {
	type Expect struct {
		boolVar   bool
		stringVar string
		intVar    int
	}

	data := []struct {
		a []string
		e Expect
	}{
		{
			a: []string{"-b", "-s", "value", "-i", "100"},
			e: Expect{true, "value", 100},
		},
		{
			a: []string{"-b=false", "-s=value", "-i=100"},
			e: Expect{false, "value", 100},
		},
		{
			a: []string{"-sib", "value", "100"},
			e: Expect{true, "value", 100},
		},
	}

	var flags *FlagSet
	for _, v := range data {
		flags = NewFlagSet("", ContinueOnError, false)
		boolVar := flags.Bool("bool", 'b', false, "", nil)
		stringVar := flags.String("string", 's', "", "", nil)
		intVar := flags.Int("int", 'i', 0, "", nil)

		if err := flags.Parse(v.a); err != nil {
			t.Errorf("should not be an error")
			break
		}
		if *boolVar != v.e.boolVar {
			t.Errorf("bool: value is should be %t", v.e.boolVar)
			break
		}
		if *stringVar != v.e.stringVar {
			t.Errorf("string: value is should be %s, got: %s", v.e.stringVar, *stringVar)
			break
		}
		if *intVar != v.e.intVar {
			t.Errorf("int: value is should be %d, got: %d", v.e.intVar, *intVar)
			break
		}
	}
}

func TestRemainderArguments(t *testing.T) {
	data := []struct {
		i bool // stop immediately if other than flag
		a []string
		r []string // remainder argument
	}{
		{
			a: []string{"-b", "A", "-s", "value", "B", "-i", "100", "C"},
			r: []string{"A", "B", "C"},
		},
		{
			a: []string{"-b=false", "-s=value", "-i=100"},
			r: []string{},
		},
		{
			a: []string{"A", "B", "C", "-sib", "value", "100"},
			r: []string{"A", "B", "C"},
		},
		{
			i: true,
			a: []string{"-b", "-s", "value", "A", "B", "-i", "100", "C"},
			r: []string{"A", "B", "-i", "100", "C"},
		},
		{
			i: true,
			a: []string{"A", "B", "C", "-sib", "value", "100"},
			r: []string{"A", "B", "C", "-sib", "value", "100"},
		},
	}

	var flags *FlagSet
	for _, v := range data {
		flags = NewFlagSet("", ContinueOnError, v.i)
		flags.Bool("bool", 'b', false, "", nil)
		flags.String("string", 's', "", "", nil)
		flags.Int("int", 'i', 0, "", nil)

		if err := flags.Parse(v.a); err != nil {
			t.Errorf("should not be an error")
			break
		}
		if !reflect.DeepEqual(v.r, flags.Args()) {
			t.Errorf("\nactual: %v\nexpect: %v\n", flags.Args(), v.r)
			break
		}
	}
}
