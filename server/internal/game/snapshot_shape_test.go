package game

import (
	"encoding"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// snapshot_shape_test.go is #522's schema-shape guard.
//
// The drift test makes every DOMAIN field say what happens to it; the
// round trip proves one binary reads what it writes. Neither notices
// the change that actually breaks a deploy: the on-disk SHAPE of the
// snapshot moving under a version number that did not. Rename a
// mirror field's JSON key, change an int to a string, or turn an
// object into a list, and yesterday's file decodes into today's binary
// as a zero value.
//
// So the shape is written down. A golden file per schema version,
// testdata/snapshot_shape/v<N>.txt, lists every JSON path GameSnapshot
// can emit and the JSON type at that path, derived by reflection over
// the snapshot structs. The rule, stated in full next to
// SnapshotSchemaVersion:
//
//   - The reflected shape must equal the golden file for the current
//     version. Any change fails here, so none of them is silent.
//   - A purely ADDITIVE change (new paths only) is recorded by
//     re-running with -update-shape, which rewrites the current
//     version's file in the same change. No bump, provided the new
//     field zero-values correctly for older files and an older binary
//     that drops it restores the game correctly — if not, bump anyway.
//   - Any other change — a path removed, renamed or retyped — needs a
//     bump. -update-shape refuses to rewrite an existing version's
//     file with one; bump SnapshotSchemaVersion and it writes the new
//     version's file, leaving the old one frozen as history.
//
//	go test ./internal/game -run TestSnapshotShapeIsRecorded -args -update-shape

var updateShape = flag.Bool("update-shape", false,
	"rewrite testdata/snapshot_shape/v<SnapshotSchemaVersion>.txt from the snapshot structs (additive changes only, or a new version)")

func shapeGoldenPath(version int) string {
	return filepath.Join("testdata", "snapshot_shape", fmt.Sprintf("v%d.txt", version))
}

// snapshotShape returns the sorted "path<TAB>type" lines for the JSON a
// GameSnapshot marshals to.
func snapshotShape() []string {
	lines := map[string]bool{}
	var walk func(rt reflect.Type, path string, stack []reflect.Type)
	emit := func(path, kind string) { lines[path+"\t"+kind] = true }
	marshaler := reflect.TypeOf((*json.Marshaler)(nil)).Elem()
	texter := reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()

	walk = func(rt reflect.Type, path string, stack []reflect.Type) {
		for rt.Kind() == reflect.Ptr {
			rt = rt.Elem()
		}
		if rt.Implements(marshaler) || reflect.PointerTo(rt).Implements(marshaler) ||
			rt.Implements(texter) || reflect.PointerTo(rt).Implements(texter) {
			emit(path, "custom("+rt.String()+")")
			return
		}
		switch rt.Kind() {
		case reflect.Bool:
			emit(path, "bool")
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			emit(path, "number:"+rt.Kind().String())
		case reflect.String:
			emit(path, "string")
		case reflect.Interface:
			emit(path, "any")
		case reflect.Slice, reflect.Array:
			if rt.Elem().Kind() == reflect.Uint8 && rt.Kind() == reflect.Slice {
				emit(path, "bytes")
				return
			}
			emit(path, "array")
			walk(rt.Elem(), path+"[]", stack)
		case reflect.Map:
			emit(path, "map:"+rt.Key().Kind().String())
			walk(rt.Elem(), path+"{}", stack)
		case reflect.Struct:
			for _, s := range stack {
				if s == rt {
					emit(path, "recursive("+rt.String()+")")
					return
				}
			}
			if path != "" {
				emit(path, "object")
			}
			walkFields(rt, path, append(stack, rt), walk)
		default:
			emit(path, "unsupported:"+rt.Kind().String())
		}
	}
	walk(reflect.TypeOf(GameSnapshot{}), "", nil)

	out := make([]string, 0, len(lines))
	for l := range lines {
		out = append(out, l)
	}
	sort.Strings(out)
	return out
}

// walkFields visits the fields encoding/json would marshal, flattening
// untagged embedded structs the way it does.
func walkFields(rt reflect.Type, path string, stack []reflect.Type,
	walk func(reflect.Type, string, []reflect.Type)) {
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		tag := f.Tag.Get("json")
		if tag == "-" {
			continue
		}
		name, _, _ := strings.Cut(tag, ",")
		if f.Anonymous && name == "" {
			ft := f.Type
			for ft.Kind() == reflect.Ptr {
				ft = ft.Elem()
			}
			if ft.Kind() == reflect.Struct {
				walkFields(ft, path, stack, walk)
				continue
			}
		}
		if !f.IsExported() {
			continue
		}
		if name == "" {
			name = f.Name
		}
		child := name
		if path != "" {
			child = path + "." + name
		}
		walk(f.Type, child, stack)
	}
}

func readShapeGolden(t *testing.T, version int) ([]string, bool) {
	t.Helper()
	raw, err := os.ReadFile(shapeGoldenPath(version))
	if os.IsNotExist(err) {
		return nil, false
	}
	if err != nil {
		t.Fatalf("read %s: %v", shapeGoldenPath(version), err)
	}
	var out []string
	for _, l := range strings.Split(string(raw), "\n") {
		if l == "" || strings.HasPrefix(l, "#") {
			continue
		}
		out = append(out, l)
	}
	return out, true
}

// shapeDiff returns the lines only in got (added) and only in want
// (removed).
func shapeDiff(want, got []string) (added, removed []string) {
	w, g := map[string]bool{}, map[string]bool{}
	for _, l := range want {
		w[l] = true
	}
	for _, l := range got {
		g[l] = true
		if !w[l] {
			added = append(added, l)
		}
	}
	for _, l := range want {
		if !g[l] {
			removed = append(removed, l)
		}
	}
	return added, removed
}

const shapeGoldenHeader = `# The on-disk shape of game.GameSnapshot at this schema version: every
# JSON path it can emit and the JSON type there, derived by reflection
# (snapshot_shape_test.go). Generated — do not edit by hand.
#
# Additive changes are recorded in the CURRENT version's file with
# -update-shape. Anything else needs a SnapshotSchemaVersion bump, and
# once the version moves on this file is frozen history (#522).
`

func TestSnapshotShapeIsRecorded(t *testing.T) {
	got := snapshotShape()
	want, ok := readShapeGolden(t, SnapshotSchemaVersion)

	if *updateShape {
		if ok {
			if _, removed := shapeDiff(want, got); len(removed) > 0 {
				t.Fatalf(`refusing to rewrite %s: the change is not additive.

These paths were removed, renamed or retyped:
  %s

A snapshot written by the binary before this change carries them, and
this binary would read them as zero values. Bump SnapshotSchemaVersion,
say why in its comment, and run -update-shape again: it will write
v%d.txt and leave v%d.txt frozen.`,
					shapeGoldenPath(SnapshotSchemaVersion), strings.Join(removed, "\n  "),
					SnapshotSchemaVersion+1, SnapshotSchemaVersion)
			}
		}
		if err := os.MkdirAll(filepath.Dir(shapeGoldenPath(SnapshotSchemaVersion)), 0o755); err != nil {
			t.Fatal(err)
		}
		body := shapeGoldenHeader + strings.Join(got, "\n") + "\n"
		if err := os.WriteFile(shapeGoldenPath(SnapshotSchemaVersion), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("wrote %s (%d paths)", shapeGoldenPath(SnapshotSchemaVersion), len(got))
		return
	}

	if !ok {
		t.Fatalf(`SnapshotSchemaVersion is %d and there is no %s.

A version bump records the new shape. Run:

  go test ./internal/game -run TestSnapshotShapeIsRecorded -args -update-shape

and write the v%d fixture set (AGENTS.md, "Snapshot compatibility").`,
			SnapshotSchemaVersion, shapeGoldenPath(SnapshotSchemaVersion), SnapshotSchemaVersion)
	}
	added, removed := shapeDiff(want, got)
	if len(removed) > 0 {
		t.Errorf(`the snapshot's on-disk shape changed in a way that is NOT additive,
under the same SnapshotSchemaVersion (%d).

Removed, renamed or retyped:
  %s

Added:
  %s

A file written before this change decodes into this binary with those
paths read as zero values, silently. That is a schema bump: raise
SnapshotSchemaVersion, say why in its comment, then run

  go test ./internal/game -run TestSnapshotShapeIsRecorded -args -update-shape

which writes v%d.txt and leaves v%d.txt frozen.`,
			SnapshotSchemaVersion, strings.Join(removed, "\n  "), strings.Join(added, "\n  "),
			SnapshotSchemaVersion+1, SnapshotSchemaVersion)
	}
	if len(added) > 0 && len(removed) == 0 {
		t.Errorf(`the snapshot's on-disk shape grew and the change is not recorded.

Added:
  %s

An additive change needs no bump IF the new field zero-values correctly
for a file written before it AND a binary before it that drops the key
still restores the game correctly (see SnapshotSchemaVersion — v2, v5
and v6 are the cases where the second half failed and forced a bump).
Record it in the same change:

  go test ./internal/game -run TestSnapshotShapeIsRecorded -args -update-shape`,
			strings.Join(added, "\n  "))
	}
}

// TestSnapshotShapeSeesARename proves the guard is not vacuous: a
// renamed JSON key on a mirror is a removed path and an added one.
func TestSnapshotShapeSeesARename(t *testing.T) {
	type before struct {
		Counters map[string]int `json:"counters"`
	}
	type after struct {
		Counters map[string]int `json:"counterMap"`
	}
	shapeOf := func(v any) []string {
		lines := map[string]bool{}
		var walk func(rt reflect.Type, path string, stack []reflect.Type)
		walk = func(rt reflect.Type, path string, stack []reflect.Type) {
			switch rt.Kind() {
			case reflect.Struct:
				walkFields(rt, path, stack, walk)
			default:
				lines[path+"\t"+rt.Kind().String()] = true
			}
		}
		walk(reflect.TypeOf(v), "", nil)
		out := []string{}
		for l := range lines {
			out = append(out, l)
		}
		sort.Strings(out)
		return out
	}
	added, removed := shapeDiff(shapeOf(before{}), shapeOf(after{}))
	if len(added) != 1 || len(removed) != 1 {
		t.Fatalf("a renamed key read as added=%v removed=%v, want one of each", added, removed)
	}
}
