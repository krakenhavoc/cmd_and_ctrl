package coverage

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/oracletext"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rider_purpose_test.go — #4978: a Purpose's riders (scry, learn,
// discover, amass and the Ring tempting its controller, ADR 0126's
// amendment of 2026-11-05) are held to the printed card, both ways.
//
//   - A declared rider is printed: a Scry 2 on a card whose text says
//     "scry 2", a Learn on one that says "learn", never more temptings
//     or learns than the text has. A declaration that the text does not
//     carry would price a card for what it does not do.
//   - A printed rider with a fixed amount is declared somewhere on the
//     card, or the card is on undeclaredRidersFile with the reason it is
//     not. The list may only shrink: a new card that prints a rider
//     declares it or names why not, and a card that comes to declare it
//     leaves the list.
//
// The text is the oracle fixture, reminder text removed. A rider counted
// at resolution ("scry X", "discover X, where X is ...") is not one this
// test reads, since no printed number would be true.

// undeclaredRidersFile pins the catalog cards that print a fixed rider
// and declare none, one "<card> | <rider> | <reason>" a line.
const undeclaredRidersFile = "testdata/undeclared_riders.txt"

// riderKind is one rider: how it is printed and where a Purpose keeps it.
type riderKind struct {
	name string
	// printed is the instruction, with the amount as the first group for
	// a rider that has one.
	printed *regexp.Regexp
	// notInstruction rejects a match that is a trigger condition or a
	// replacement rather than the instruction ("Whenever the Ring tempts
	// you", "if you would learn"): the text just before it.
	notInstruction *regexp.Regexp
	declared       func(game.Purpose) int
	// counted: the declaration is a count of the instruction (learn, the
	// Ring), not its printed amount.
	counted bool
}

var riderKinds = []riderKind{
	{name: "scry", printed: regexp.MustCompile(`\bscry (\d+)\b`), declared: func(p game.Purpose) int { return p.Scry }},
	{name: "learn", printed: regexp.MustCompile(`\blearn\b`), notInstruction: regexp.MustCompile(`(would|whenever you|if you) $`),
		declared: func(p game.Purpose) int { return p.Learn }, counted: true},
	{name: "discover", printed: regexp.MustCompile(`\bdiscover (\d+)\b`), declared: func(p game.Purpose) int { return p.Discover }},
	{name: "amass", printed: regexp.MustCompile(`\bamass [a-z]+ (\d+)\b`), declared: func(p game.Purpose) int { return p.Amass }},
	{name: "ring", printed: regexp.MustCompile(`\bthe ring tempts you\b`), notInstruction: regexp.MustCompile(`(whenever|when|each time|time|times) $`),
		declared: func(p game.Purpose) int { return p.RingTempts }, counted: true},
}

// printedRiders is, per rider kind, the amounts (or for a counted kind,
// one entry per instruction) the card's text prints.
func printedRiders(c OracleCard) map[string][]int {
	out := map[string][]int{}
	for _, f := range c.AllFaces() {
		for _, line := range strings.Split(f.Text, "\n") {
			text := strings.ToLower(oracletext.StripReminder(line))
			for _, k := range riderKinds {
				for _, m := range k.printed.FindAllStringSubmatchIndex(text, -1) {
					if k.notInstruction != nil && k.notInstruction.MatchString(text[:m[0]]) {
						continue
					}
					n := 1
					if len(m) >= 4 && m[2] >= 0 {
						n, _ = strconv.Atoi(text[m[2]:m[3]])
					}
					out[k.name] = append(out[k.name], n)
				}
			}
		}
	}
	return out
}

// specPurposes is every Purpose a Spec declares, with where.
func specPurposes(s effects.Spec) map[string]game.Purpose {
	out := map[string]game.Purpose{}
	var addPurpose func(where string, p game.Purpose)
	addPurpose = func(where string, p game.Purpose) {
		out[where] = p
		if p.Kicked != nil {
			addPurpose(where+" kicked", *p.Kicked)
		}
		if p.FaceUp != nil {
			addPurpose(where+" face up", *p.FaceUp)
		}
	}
	addModes := func(where string, m *game.ModeSpec) {
		if m == nil {
			return
		}
		for i, o := range m.Options {
			addPurpose(fmt.Sprintf("%s mode %d", where, i), o.Purpose)
		}
	}
	addPurpose("card", s.Purpose)
	addModes("card", s.Modes)
	for i, a := range s.AlternativeCosts {
		addPurpose(fmt.Sprintf("alternative cost %d", i), a.Purpose)
	}
	for i, a := range s.Activated {
		addPurpose(fmt.Sprintf("ability %d", i), a.Purpose)
		addModes(fmt.Sprintf("ability %d", i), a.Modes)
	}
	for i, t := range s.Triggered {
		addPurpose(fmt.Sprintf("triggered ability %d", i), t.Purpose)
		addModes(fmt.Sprintf("triggered ability %d", i), t.Modes)
	}
	return out
}

func loadUndeclaredRiders(t *testing.T) map[string]string {
	t.Helper()
	f, err := os.Open(undeclaredRidersFile)
	if err != nil {
		t.Fatalf("open %s: %v", undeclaredRidersFile, err)
	}
	defer func() { _ = f.Close() }()
	out := map[string]string{}
	sc := bufio.NewScanner(f)
	for n := 1; sc.Scan(); n++ {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, " | ")
		if len(parts) != 3 || strings.TrimSpace(parts[2]) == "" {
			t.Fatalf("%s:%d: want \"<card> | <rider> | <reason>\", got %q", undeclaredRidersFile, n, line)
		}
		key := parts[0] + " | " + parts[1]
		if _, dup := out[key]; dup {
			t.Fatalf("%s:%d: %q is listed twice", undeclaredRidersFile, n, key)
		}
		out[key] = parts[2]
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("read %s: %v", undeclaredRidersFile, err)
	}
	return out
}

func TestRiderPurposesMatchThePrintedCard(t *testing.T) {
	fx := loadOracleFixture(t)
	// Every Spec a card registers, face 0 and face 1 under one base ID.
	byCard := map[string][]effects.Spec{}
	for _, s := range effects.All() {
		base, _, _ := strings.Cut(s.OracleID, "#")
		byCard[base] = append(byCard[base], s)
	}
	pinned := loadUndeclaredRiders(t)
	var wrong []string
	undeclared := map[string]bool{}
	declaredAny := 0
	for id, specs := range byCard {
		card, ok := fx[id]
		if !ok {
			continue
		}
		printed := printedRiders(card)
		for _, k := range riderKinds {
			total := 0
			for _, s := range specs {
				for where, p := range specPurposes(s) {
					n := k.declared(p)
					if n == 0 {
						continue
					}
					total += n
					declaredAny++
					amounts := printed[k.name]
					switch {
					case len(amounts) == 0:
						wrong = append(wrong, fmt.Sprintf("%s %s declares %s %d, which the card does not print", card.Name, where, k.name, n))
					case !k.counted && !containsInt(amounts, n) && n != sum(amounts):
						wrong = append(wrong, fmt.Sprintf("%s %s declares %s %d; the card prints %s %v", card.Name, where, k.name, n, k.name, amounts))
					}
				}
			}
			if k.counted && total > len(printed[k.name]) {
				wrong = append(wrong, fmt.Sprintf("%s declares %s %d times; the card prints it %d", card.Name, k.name, total, len(printed[k.name])))
			}
			if total == 0 && len(printed[k.name]) > 0 {
				undeclared[card.Name+" | "+k.name] = true
			}
		}
	}
	if declaredAny == 0 {
		t.Fatal("no catalog card declares a rider — a guard that reads nothing passes everything")
	}
	sort.Strings(wrong)
	for _, w := range wrong {
		t.Error(w)
	}
	var added, gone []string
	for key := range undeclared {
		if _, ok := pinned[key]; !ok {
			added = append(added, key)
		}
	}
	for key := range pinned {
		if !undeclared[key] {
			gone = append(gone, key)
		}
	}
	sort.Strings(added)
	sort.Strings(gone)
	if len(added) > 0 {
		t.Errorf("these cards print a rider with a fixed amount and declare none (#4978): declare it on the slot that does it "+
			"(the card for a spell or an enters effect, the row, the bullet), or add a line with the reason to %s:\n  %s",
			undeclaredRidersFile, strings.Join(added, "\n  "))
	}
	if len(gone) > 0 {
		t.Errorf("these lines of %s no longer match an undeclared rider; delete them:\n  %s",
			undeclaredRidersFile, strings.Join(gone, "\n  "))
	}
}

func containsInt(xs []int, n int) bool {
	for _, x := range xs {
		if x == n {
			return true
		}
	}
	return false
}

func sum(xs []int) int {
	s := 0
	for _, x := range xs {
		s += x
	}
	return s
}

// The text reader finds the instruction and not the trigger condition
// or the replacement.
func TestPrintedRidersReadInstructionsOnly(t *testing.T) {
	for text, want := range map[string]map[string][]int{
		"Scry 2, then draw a card.":                                                                 {"scry": {2}},
		"Destroy target creature. The Ring tempts you.":                                             {"ring": {1}},
		"Whenever the Ring tempts you, draw a card.":                                                {},
		"When this creature enters, amass Orcs 2. (Put two +1/+1 counters on.)":                     {"amass": {2}},
		"As long as this card is in your graveyard, if you would learn, you may instead return it.": {},
		"Learn. (You may reveal a Lesson card you own from outside the game.)":                      {"learn": {1}},
		"Discover 4.":                       {"discover": {4}},
		"Discover X, where X is its power.": {},
	} {
		got := printedRiders(OracleCard{Name: "Probe", Text: text})
		if fmt.Sprint(got) != fmt.Sprint(want) {
			t.Errorf("printedRiders(%q) = %v, want %v", text, got, want)
		}
	}
}
