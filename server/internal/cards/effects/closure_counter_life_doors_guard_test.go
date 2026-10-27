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

// closure_counter_life_doors_guard_test.go — #4253 slice 7: the counter and
// life-change doors that take a Go closure (AddCounterByThenForEffect,
// ChangePlayerLifeThenForEffect and the earthbend closure form; #4253's last
// part deleted AddCounterThenForEffect, LoseLifeEachThenForEffect and
// AmassForEffect) are called by NO catalog card. A table waiting
// on a CR 616 prompt over a counter placement or a life change whose "then" is a
// closure is not a restore point, so a card passes a registered tail to the
// Keyed form instead (AddCounterByKeyedForEffect, ChangePlayerLifeKeyedForEffect,
// LoseLifeEachKeyedForEffect, AmassKeyedForEffect, EarthbendKeyedForEffect).
//
// The closure doors stay for the engine's own plumbing and for tests that
// record what a continuation was told. This test counts every other production
// call, by file, against the list below. THE LIST MAY ONLY FALL: moving a
// caller onto a tail takes its line out; adding a closure caller is refused.

// closureDoorCalls is every non-test call of a closure door still in the
// tree, by file under server/internal, with why it is there.
var closureDoorCalls = map[string]struct {
	n   int
	why string
}{
	"game/effect_api.go":  {2, "ChangePlayerLifeForEffect and AddCounterForEffect pass a nil then"},
	"game/earthbend.go":   {1, "applyEarthbendLocked's closure form, reached only through EarthbendThenForEffect with a closure"},
	"game/proliferate.go": {1, "the proliferate counter loop: its rest is still a closure (the ask prompt's chooseCardsFrame.then, ProliferateChoosingForEffect)"},
}

var closureDoorCall = regexp.MustCompile(`\b(AddCounterThenForEffect|AddCounterByThenForEffect|ChangePlayerLifeThenForEffect|LoseLifeEachThenForEffect|AmassForEffect|EarthbendThenForEffect)\(`)

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
