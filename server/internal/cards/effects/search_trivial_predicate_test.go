package effects

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestNoSearchPassesATrivialPredicate guards #3053's one card-author
// trap. The engine reads a search's Predicate as "this search names a
// quality" (CR 701.23b: the searcher may find nothing); a search for
// "a card" names none (CR 701.23d: the searcher must find one) and says
// so by OMITTING the predicate. `func(game.Card) bool { return true }`
// looks the same and is not: it turns Demonic Tutor into a tutor that
// may fail to find. Seven catalog cards did exactly that.
func TestNoSearchPassesATrivialPredicate(t *testing.T) {
	trivial := regexp.MustCompile(`(?:Predicate|Pred):\s*func\([^)]*\)\s*bool\s*\{\s*return true\s*\}`)
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if loc := trivial.FindIndex(src); loc != nil {
			line := 1 + strings.Count(string(src[:loc[0]]), "\n")
			t.Errorf("%s:%d: a search passes an always-true predicate; omit Predicate for an unconditional \"a card\" search (CR 701.23d)", f, line)
		}
	}
}
