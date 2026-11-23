package main

// overrides.go — the reviewed rename table (overrides.txt).
//
// By default an unexported name that has to be exported is upper-cased
// (onBattlefield → OnBattlefield). When that collides with a name already in
// the package (or with a local at a use, or with a method or field of the
// same type), overrides.txt says what to call it instead. The tool never
// invents a name silently: a collision with no override gets an X prefix
// and is listed on stderr, so it can be reviewed and added here.

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type overrides struct {
	name   map[string]string // package-level object, by old name
	method map[string]string // every method of that name
	field  map[string]string // <OldTypeName>.<field>
}

func readOverrides(path string) (*overrides, error) {
	ov := &overrides{name: map[string]string{}, method: map[string]string{}, field: map[string]string{}}
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return ov, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	n := 0
	for sc.Scan() {
		n++
		line := sc.Text()
		if i := strings.Index(line, "#"); i >= 0 {
			line = line[:i]
		}
		fs := strings.Fields(line)
		if len(fs) == 0 {
			continue
		}
		if len(fs) != 3 {
			return nil, fmt.Errorf("%s:%d: want <name|method|field> <old> <new>", path, n)
		}
		var m map[string]string
		switch fs[0] {
		case "name":
			m = ov.name
		case "method":
			m = ov.method
		case "field":
			m = ov.field
		default:
			return nil, fmt.Errorf("%s:%d: unknown kind %q", path, n, fs[0])
		}
		if _, dup := m[fs[1]]; dup {
			return nil, fmt.Errorf("%s:%d: %s %s is listed twice", path, n, fs[0], fs[1])
		}
		m[fs[1]] = fs[2]
	}
	return ov, sc.Err()
}
