package coverage

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/oracletext"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trigger_footprint_audit_test.go — ADR 0018's #3013 amendment, sub-PR 1.
//
// "Ask only when the order matters" can skip a CR 603.3b prompt only for
// triggers whose rows carry a footprint (#2884), and a row gets one only
// when its effect is an effects.Do of primitives that declare a step. In
// October 2026 most one-sentence triggers ("…, put a +1/+1 counter on
// ~", "…, you gain 2 life") were hand-written closures, so the batches
// they make keep asking although every order gives the same game.
//
// This audit reads the printed text from the oracle fixtures (the
// Scryfall projection CI already holds, refreshed against the dump by
// TestOracleFixtureIsCurrent every night). A catalogued card is LISTED
// when it prints more one-sentence trigger lines of a shape the
// vocabulary has, or is one primitive away from, than it has triggered
// rows with a footprint. Every listed card is in
// testdata/unfootprinted_trigger_lines.txt, and the file ONLY SHRINKS:
//
//   - a listed card that no longer needs listing (it was moved onto Do
//     primitives) fails until its line is deleted;
//   - a card that needs listing and is not listed (a NEW one-sentence
//     trigger written as a closure) fails: write it with Do instead.
//
// Delete the stale lines with
//
//	go test ./internal/cards/coverage -run TestOneSentenceTriggersDeclareFootprints -args -update-unfootprinted
//
// which never adds a line. A shape the vocabulary cannot express yet is
// still listed: that is the backlog the amendment's sub-PRs work through.

var updateUnfootprinted = flag.Bool("update-unfootprinted", false,
	"delete stale lines from testdata/unfootprinted_trigger_lines.txt (never adds one)")

const unfootprintedTriggerLinesFile = "testdata/unfootprinted_trigger_lines.txt"

// oneSentenceTriggerShapes are the one-sentence trigger effects the
// footprint vocabulary has (game.FootprintKind) or is one primitive away
// from, matched against a normalize()d line: lower case, the card's
// name as "~", no final period. The trigger condition is anything up to
// the last ", " before the effect.
var oneSentenceTriggerShapes = func() map[string]*regexp.Regexp {
	const (
		cond = `^(when|whenever|at the beginning of)[^.;"]*, `
		num  = `(a|an|one|two|three|four|five|\d+)`
	)
	return map[string]*regexp.Regexp{
		"counter on source":          regexp.MustCompile(cond + `put ` + num + ` \+1/\+1 counters? on ~$`),
		"gain life":                  regexp.MustCompile(cond + `you gain ` + num + ` life$`),
		"gain life equal to subject": regexp.MustCompile(cond + `you gain life equal to (its|that creature's) (toughness|power)$`),
		"draw":                       regexp.MustCompile(cond + `draw ` + num + ` cards?$`),
		"mill, scry or surveil":      regexp.MustCompile(cond + `(mill|scry|surveil) ` + num + `( cards?)?$`),
		"create tokens":              regexp.MustCompile(cond + `create ` + num + ` [^"]*?tokens?$`),
		"drain":                      regexp.MustCompile(cond + `each opponent loses ` + num + ` life( and you gain ` + num + ` life)?$`),
		"damage each opponent":       regexp.MustCompile(cond + `~ deals ` + num + ` damage to each opponent$`),
		"energy":                     regexp.MustCompile(cond + `you get (\{e\})+$`),
		"monarch":                    regexp.MustCompile(cond + `you become the monarch$`),
	}
}()

// oneSentenceShape names the shape a normalized line has, or "".
func oneSentenceShape(norm string) string {
	for name, re := range oneSentenceTriggerShapes {
		if re.MatchString(norm) {
			return name
		}
	}
	return ""
}

// specTriggerText is the printed text a Spec key's triggers come from: a
// back face's own text for "<id>#n", and for a base key the front face
// when a back face is registered separately, else every face.
func specTriggerText(key, name string, card OracleCard, registered map[string]bool) []string {
	base, face := BaseOracleID(key)
	faces := card.AllFaces()
	names := cardNames(name, card)
	var texts []string
	switch {
	case face > 0 && face < len(faces):
		texts = []string{faces[face].Text}
	case face > 0:
		return nil
	case len(faces) > 1 && registered[base+"#1"]:
		texts = []string{faces[0].Text}
	default:
		for _, f := range faces {
			texts = append(texts, f.Text)
		}
	}
	var out []string
	for _, text := range texts {
		for _, l := range abilityLines(text, names) {
			out = append(out, l.norm)
		}
	}
	return out
}

// unfootprintedLine is one card's line in the list.
func unfootprintedLine(name, key string) string { return name + " | " + key }

// wantUnfootprintedLines is the set the file must equal, with each
// line's shapes for the failure message.
func wantUnfootprintedLines(t *testing.T) map[string][]string {
	t.Helper()
	fx := loadOracleFixture(t)
	specs := effects.All()
	registered := map[string]bool{}
	for _, s := range specs {
		registered[s.OracleID] = true
	}
	want := map[string][]string{}
	for _, s := range specs {
		if len(s.Triggered) == 0 {
			continue
		}
		base, _ := BaseOracleID(s.OracleID)
		card, ok := fx[base]
		if !ok {
			continue
		}
		var shapes []string
		for _, line := range specTriggerText(s.OracleID, s.Name, card, registered) {
			if shape := oneSentenceShape(line); shape != "" {
				shapes = append(shapes, shape)
			}
		}
		if len(shapes) == 0 {
			continue
		}
		footprinted := 0
		for _, row := range game.CatalogTriggers(s.OracleID) {
			if len(row.Footprint) > 0 {
				footprinted++
			}
		}
		if len(shapes) > footprinted {
			sort.Strings(shapes)
			want[unfootprintedLine(s.Name, s.OracleID)] = shapes
		}
	}
	return want
}

func readUnfootprintedLines(t *testing.T) ([]string, bool) {
	t.Helper()
	raw, err := os.ReadFile(unfootprintedTriggerLinesFile)
	if os.IsNotExist(err) {
		return nil, false
	}
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		out = append(out, line)
	}
	return out, true
}

const unfootprintedHeader = `# unfootprinted_trigger_lines.txt — ADR 0018's #3013 amendment, sub-PR 1.
#
# Every catalogued card that prints more one-sentence trigger lines of a
# shape the footprint vocabulary has (or is one primitive away from) than
# it has triggered rows with a footprint. Their batches keep the CR 603.3b
# prompt under "Ask only when the order matters" (#2884).
#
# One card per line, sorted: <card name> | <catalog key>.
# The list ONLY SHRINKS. Move a card's trigger onto effects.Do primitives
# that declare a step (cards/effects/footprint.go) and delete its line, or run
#   go test ./internal/cards/coverage -run TestOneSentenceTriggersDeclareFootprints -args -update-unfootprinted
# which deletes stale lines and never adds one. A new one-sentence trigger
# written as a closure fails the test instead of being added here.
`

func writeUnfootprintedLines(t *testing.T, lines []string) {
	t.Helper()
	sort.Strings(lines)
	var b strings.Builder
	b.WriteString(unfootprintedHeader)
	b.WriteString("\n")
	for _, l := range lines {
		b.WriteString(l)
		b.WriteString("\n")
	}
	if err := os.WriteFile(unfootprintedTriggerLinesFile, []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestOneSentenceTriggersDeclareFootprints(t *testing.T) {
	want := wantUnfootprintedLines(t)
	have, exists := readUnfootprintedLines(t)

	if *updateUnfootprinted {
		var keep []string
		if !exists {
			// Bootstrapping is the one time the list is written whole:
			// the change that created it.
			for l := range want {
				keep = append(keep, l)
			}
		} else {
			for _, l := range have {
				if _, ok := want[l]; ok {
					keep = append(keep, l)
				}
			}
		}
		writeUnfootprintedLines(t, keep)
		have, exists = readUnfootprintedLines(t)
	}
	if !exists {
		t.Fatalf("%s is missing", unfootprintedTriggerLinesFile)
	}

	listed := map[string]bool{}
	var problems []string
	for i, l := range have {
		if listed[l] {
			problems = append(problems, "listed twice: "+l)
		}
		listed[l] = true
		if i > 0 && have[i-1] > l {
			problems = append(problems, fmt.Sprintf("not sorted: %q comes after %q", l, have[i-1]))
		}
		if _, ok := want[l]; !ok {
			problems = append(problems, "no longer needs listing — delete the line (the list only shrinks): "+l)
		}
	}
	var missing []string
	for l := range want {
		if !listed[l] {
			missing = append(missing, l)
		}
	}
	sort.Strings(missing)
	for _, l := range missing {
		problems = append(problems, fmt.Sprintf(
			"prints a one-sentence trigger (%s) whose row has no footprint, and is not listed — write the effect with effects.Do primitives that declare a step (cards/effects/footprint.go) instead: %s",
			strings.Join(want[l], ", "), l))
	}
	if len(problems) > 0 {
		t.Errorf("%s:\n  %s", unfootprintedTriggerLinesFile, strings.Join(problems, "\n  "))
	}
	t.Logf("%d cards listed", len(have))
}

// The audit is not vacuous: each shape matches the line it is for, and
// the cards the amendment names are read the way it says.
func TestOneSentenceTriggerShapes(t *testing.T) {
	for _, c := range []struct {
		line, shape string
	}{
		{"Whenever another Elf you control enters, put a +1/+1 counter on Marwyn.", "counter on source"},
		{"Whenever another Elf you control enters, you gain life equal to its toughness.", "gain life equal to subject"},
		{"Whenever another creature enters, you gain 1 life.", "gain life"},
		{"Whenever a creature you control dies, each opponent loses 1 life and you gain 1 life.", "drain"},
		{"At the beginning of each upkeep, create a 1/1 green Elf Warrior creature token.", "create tokens"},
		{"Whenever you cast a noncreature spell, Guttersnipe deals 2 damage to each opponent.", "damage each opponent"},
		{"Whenever you cast a spell, you may draw a card.", ""},
		{"Whenever this creature attacks, draw a card, then discard a card.", ""},
	} {
		names := []string{"Marwyn, the Nurturer", "Marwyn", "Guttersnipe"}
		if got := oneSentenceShape(oracletext.Normalize(c.line, names)); got != c.shape {
			t.Errorf("%q: shape %q, want %q", c.line, got, c.shape)
		}
	}
}
