package effects

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// closure_counter_life_doors_guard_test.go — #4253 slice 7: no catalog card
// calls a life-change or counter door that takes a Go closure. #4253 deleted
// AddCounterByThenForEffect, AddCounterThenForEffect, LoseLifeEachThenForEffect,
// AmassForEffect and the earthbend closure form, and #4361 S9 deleted the last,
// ChangePlayerLifeThenForEffect; this guard keeps the name from coming back. A table waiting
// on a CR 616 prompt over a counter placement or a life change whose "then" is a
// closure is not a restore point, so a card passes a registered tail to the
// Keyed form instead (AddCounterByKeyedForEffect, ChangePlayerLifeKeyedForEffect,
// LoseLifeEachKeyedForEffect, AmassKeyedForEffect, EarthbendKeyedForEffect).
//
// This test counts every production call, by file, against the list below,
// which is empty now. THE LIST MAY ONLY FALL: adding a closure caller is
// refused.

// closureDoorCalls is every non-test call of a closure door still in the
// tree, by file under server/internal, with why it is there.
var closureDoorCalls = map[string]struct {
	n   int
	why string
}{}

var closureDoorCall = regexp.MustCompile(`\b(ChangePlayerLifeThenForEffect)\(`)

func TestNoCatalogCardCallsAClosureCounterOrLifeDoor(t *testing.T) {
	root := filepath.Join("..", "..")
	got := map[string]int{}
	scanned := 0
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		scanned++
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		sc := bufio.NewScanner(f)
		sc.Buffer(make([]byte, 1<<20), 1<<20)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if strings.HasPrefix(line, "//") || strings.HasPrefix(line, "func (g *Game)") {
				continue
			}
			if closureDoorCall.MatchString(line) {
				got[rel]++
			}
		}
		return sc.Err()
	})
	if err != nil {
		t.Fatal(err)
	}
	if scanned < 100 {
		t.Fatalf("scanned %d files; the walk is wrong, and a guard that reads nothing passes everything", scanned)
	}
	var problems []string
	for file, n := range got {
		want := closureDoorCalls[file].n
		if n > want {
			problems = append(problems, file+": "+strconv.Itoa(n)+" closure counter/life calls, "+strconv.Itoa(want)+" allowed; "+
				"pass a registered tail to the Keyed form (game.RegisterTail, effects.CardTail)")
		}
	}
	for file, e := range closureDoorCalls {
		if got[file] < e.n {
			problems = append(problems, file+": "+strconv.Itoa(got[file])+" closure calls, the list says "+strconv.Itoa(e.n)+
				" ("+e.why+"); the list may only fall, so lower it")
		}
	}
	if len(problems) > 0 {
		sort.Strings(problems)
		t.Errorf("closure counter/life doors (#4253 slice 7):\n  %s", strings.Join(problems, "\n  "))
	}
}
