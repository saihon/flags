package flags

import (
	"fmt"
	"os"
	"strconv"
)

type versionValue struct {
	b       *bool
	version string
}

func newVersionValue(p *bool, version string) *versionValue {
	v := &versionValue{b: p, version: version}
	if v.b == nil {
		v.b = new(bool)
	}
	return v
}

func (v *versionValue) Set(s string) error {
	b, err := strconv.ParseBool(s)
	if err != nil {
		err = errParse
	} else {
		if b && CommandLine.errorHandling == ExitOnError {
			name := CommandLine.Name()
			if name == "" {
				name = os.Args[0]
			}
			fmt.Fprintf(CommandLine.Output(), "%s: %s\n", name, v.version)
			return ErrHelp
		}
	}
	*v.b = b
	return err
}

func (v *versionValue) Get() interface{} { return v.b }

func (v *versionValue) String() string { return strconv.FormatBool(*v.b) }

func (v *versionValue) IsBoolFlag() bool { return true }

func (f *FlagSet) VersionVar(p *bool, name string, alias rune, version string, usage string) {
	v := newVersionValue(p, version)
	f.Var(v, name, alias, usage)
}

func VersionVar(p *bool, name string, alias rune, version string, usage string) {
	v := newVersionValue(p, version)
	CommandLine.Var(v, name, alias, usage)
}

func (f *FlagSet) Version(name string, alias rune, version string, usage string) *bool {
	v := newVersionValue(nil, version)
	f.Var(v, name, alias, usage)
	return v.b
}

func Version(name string, alias rune, version string, usage string) *bool {
	return CommandLine.Version(name, alias, version, usage)
}
