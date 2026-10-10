package effects

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// answers_fallback_test.go — ADR 0142 §4: the ratchet on the
// printed-text read.
//
// Smart autopass asks legal.AnswersOf about every untargeted
// instant-speed activation. A catalog row that declares
// Purpose.Answers is read from its declaration; one that does not falls
// back to the #2853 text read. Every catalog row in scope that still
// declares nothing is listed in testdata/answers_fallback.txt, one per
// line, sorted:
//
//	<oracle_id>\t<card name>\t<ref>\t<label>
//
// <ref> is own:<i> for a card's own row, grant:<bundle>:<i> for a row of
// an ADR 0093 bundle, token:<slug>:<i> for a token's, and mana:<i> (or
// token:<slug>:mana:<i>) for a mana ability with a sacrifice-another
// cost. A token's oracle column is "token".
//
// The list ONLY SHRINKS. The test fails
//
//   - on a row in scope that is neither declared nor listed: a new
//     card's row declares its Answers (docs/adding-cards.md, "Declaring
//     what an ability answers");
//   - on a listed row that is now declared or gone: delete the line, or
//     run
//
//     go test ./internal/cards/effects -run TestAnswersFallbackOnlyShrinks -args -update-answers-fallback
//
//     which only ever removes lines;
//   - on a line listed twice, or out of order;
//   - on a list longer than answersFallbackCeiling, which each sweep PR
//     lowers to the new count and nothing raises.
//
// testdata/answers_disagreements.txt is the review record: every
// declared row whose declared tiers differ from what the text read would
// have said. TestAnswersDisagreementsAreCurrent fails when it is stale;
// regenerate it with -update-answers-disagreements, and read its diff in
// review — it is the list of verdicts the batch changed.

var (
	updateAnswersFallback = flag.Bool("update-answers-fallback", false,
		"delete stale lines from testdata/answers_fallback.txt (never adds one, except when bootstrapping a missing file)")
	updateAnswersDisagreements = flag.Bool("update-answers-disagreements", false,
		"rewrite testdata/answers_disagreements.txt from the catalog")
)

const (
	answersFallbackFile      = "testdata/answers_fallback.txt"
	answersDisagreementsFile = "testdata/answers_disagreements.txt"

	// answersFallbackCeiling is the longest the fallback list may be.
	// The signal PR bootstrapped it at the measured count; each sweep PR
	// lowers it to its new length, and nothing raises it. S4 makes it 0.
	answersFallbackCeiling = 171
)

// answersRow is one catalog row the enumerator asks answersOf (or, for
// a mana ability, manaAnswersOf) about.
type answersRow struct {
	oracle, name, ref, label string
	// Exactly one of activated and mana is set.
	activated *game.ActivatedAbilityShape
	mana      *game.ManaAbilityShape
}

func (r answersRow) line() string {
	clean := func(s string) string {
		return strings.Join(strings.Fields(strings.NewReplacer("\t", " ", "\n", " ").Replace(s)), " ")
	}
	return r.oracle + "\t" + clean(r.name) + "\t" + r.ref + "\t" + clean(r.label)
}

func (r answersRow) declared() game.Answers {
	if r.mana != nil {
		return r.mana.Answers
	}
	return r.activated.Purpose.Answers
}

// fallback is what the text (or cost) read says of the row, declared or
// not.
func (r answersRow) fallback() game.Answers {
	if r.mana != nil {
		return legal.ManaCostAnswers(r.mana.SacrificeOther)
	}
	return legal.FallbackAnswers(*r.activated)
}

// answersKeyOwner is who declares a catalog key's rows: the oracle
// column, the name, and the ref prefix.
type answersKeyOwner struct {
	oracle, name, prefix string
}

func answersKeyOwners() map[string]answersKeyOwner {
	out := map[string]answersKeyOwner{}
	for id, spec := range registry {
		out[id] = answersKeyOwner{oracle: id, name: spec.Name, prefix: "own"}
		if spec.Emblem != nil {
			out[game.EmblemKey(id)] = answersKeyOwner{oracle: id, name: spec.Name + " emblem", prefix: "emblem"}
		}
		for _, gr := range spec.Grants {
			out[game.GrantKey(gr.Key)] = answersKeyOwner{oracle: id, name: spec.Name, prefix: "grant:" + gr.Key}
		}
	}
	for slug, build := range tokenTemplatesBySlug {
		t := build()
		out[game.TokenKey(slug)] = answersKeyOwner{oracle: "token", name: t.Card.Name + " token", prefix: "token:" + slug}
		for _, gr := range t.Grants {
			out[game.GrantKey(gr.Key)] = answersKeyOwner{oracle: "token", name: t.Card.Name + " token", prefix: "grant:" + gr.Key}
		}
	}
	return out
}

// answersRows is every production catalog row in scope.
func answersRows(t *testing.T) []answersRow {
	t.Helper()
	if len(productionDefKeys) == 0 {
		t.Fatal("TestMain did not capture the production catalog keys")
	}
	owners := answersKeyOwners()
	var rows []answersRow
	for _, key := range productionDefKeys {
		d := defs[key]
		if d == nil {
			continue
		}
		o, ok := owners[key]
		if !ok {
			o = answersKeyOwner{oracle: key, name: key, prefix: "own"}
		}
		for i := range d.Activated {
			ab := d.Activated[i]
			if !answersInScope(ab) {
				continue
			}
			rows = append(rows, answersRow{
				oracle: o.oracle, name: o.name, ref: o.prefix + ":" + strconv.Itoa(i), label: ab.Label, activated: &ab,
			})
		}
		manaPrefix := "mana"
		if o.prefix != "own" {
			manaPrefix = o.prefix + ":mana"
		}
		for i := range d.ManaAbilities {
			m := d.ManaAbilities[i]
			if m.SacrificeOther == nil {
				continue
			}
			rows = append(rows, answersRow{
				oracle: o.oracle, name: o.name, ref: manaPrefix + ":" + strconv.Itoa(i), label: m.Label, mana: &m,
			})
		}
	}
	return rows
}

func readAnswersLines(t *testing.T, path string) ([]string, bool) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, false
	}
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, line := range strings.Split(string(raw), "\n") {
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue
		}
		out = append(out, line)
	}
	return out, true
}

func writeAnswersLines(t *testing.T, path, header string, lines []string) {
	t.Helper()
	sort.Strings(lines)
	var b strings.Builder
	b.WriteString(header)
	b.WriteString("\n")
	for _, l := range lines {
		b.WriteString(l)
		b.WriteString("\n")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
}

const answersFallbackHeader = `# answers_fallback.txt — ADR 0142 §4's ratchet (#2872).
#
# Every catalog row smart autopass asks about (an instant-speed
# activated row that can be announced with no target, or a mana ability
# that sacrifices another permanent) that does not declare
# Purpose.Answers (ManaAbility.Answers), so it is still read from its
# printed text and cost.
#
# One row per line, tab-separated, sorted: <oracle_id> <card name> <ref> <label>.
# The list ONLY SHRINKS. Declare the row's Answers (docs/adding-cards.md,
# "Declaring what an ability answers") and delete its line, or run
#   go test ./internal/cards/effects -run TestAnswersFallbackOnlyShrinks -args -update-answers-fallback
# which deletes stale lines and never adds one. A new undeclared row
# fails the test instead of being added here.
`

func TestAnswersFallbackOnlyShrinks(t *testing.T) {
	want := map[string]bool{}
	for _, r := range answersRows(t) {
		if r.declared() == 0 {
			want[r.line()] = true
		}
	}
	have, exists := readAnswersLines(t, answersFallbackFile)

	if *updateAnswersFallback {
		var keep []string
		if !exists {
			// Bootstrapping the list is the one time it is written
			// whole: the signal PR that created it.
			for l := range want {
				keep = append(keep, l)
			}
		} else {
			for _, l := range have {
				if want[l] {
					keep = append(keep, l)
				}
			}
		}
		writeAnswersLines(t, answersFallbackFile, answersFallbackHeader, keep)
		have, exists = readAnswersLines(t, answersFallbackFile)
	}
	if !exists {
		t.Fatalf("%s is missing", answersFallbackFile)
	}

	listed := map[string]bool{}
	var problems []string
	for i, l := range have {
		if listed[l] {
			problems = append(problems, fmt.Sprintf("listed twice: %s", l))
		}
		listed[l] = true
		if i > 0 && have[i-1] > l {
			problems = append(problems, fmt.Sprintf("not sorted: %q comes after %q", l, have[i-1]))
		}
		if !want[l] {
			problems = append(problems, fmt.Sprintf("declared now, or gone — delete the line (the list only shrinks): %s", l))
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
		problems = append(problems, fmt.Sprintf("a row smart autopass reads declares no Answers and is not listed — declare Purpose.Answers on it (docs/adding-cards.md, \"Declaring what an ability answers\", ADR 0142): %s", l))
	}
	if len(have) > answersFallbackCeiling {
		problems = append(problems, fmt.Sprintf("%d lines is over answersFallbackCeiling (%d): the list only shrinks", len(have), answersFallbackCeiling))
	}
	if len(problems) > 0 {
		t.Errorf("%s:\n  %s", answersFallbackFile, strings.Join(problems, "\n  "))
	}
	t.Logf("%d rows in scope, %d still on the fallback (ceiling %d)", len(answersRows(t)), len(have), answersFallbackCeiling)
}

const answersDisagreementsHeader = `# answers_disagreements.txt — ADR 0142 §4's review record (#2872).
#
# Every catalog row that declares its Answers where the declared tiers
# differ from what the printed-text read (legal.FallbackAnswers) would
# have said: the verdicts the declarations changed, a pointless stop
# removed or a missed window closed. Generated, never edited by hand:
#   go test ./internal/cards/effects -run TestAnswersDisagreementsAreCurrent -args -update-answers-disagreements
# Read its diff in review.
#
# One row per line, tab-separated, sorted:
# <oracle_id> <card name> <ref> <label> declared=<answers> (<tiers>) fallback=<answers> (<tiers>)
`

// answersTiers names the tiers a set reaches, for the review record.
func answersTiers(a game.Answers) string {
	var ts []string
	if a.HasTier(game.TierStack) {
		ts = append(ts, "stack")
	}
	if a.HasTier(game.TierCombat) {
		ts = append(ts, "combat")
	}
	if len(ts) == 0 {
		return "none"
	}
	return strings.Join(ts, "+")
}

func wantAnswersDisagreements(t *testing.T) []string {
	t.Helper()
	var out []string
	for _, r := range answersRows(t) {
		d := r.declared()
		if d == 0 {
			continue
		}
		f := r.fallback()
		if answersTiers(d) == answersTiers(f) {
			continue
		}
		out = append(out, fmt.Sprintf("%s\tdeclared=%s (%s)\tfallback=%s (%s)", r.line(), d, answersTiers(d), f, answersTiers(f)))
	}
	sort.Strings(out)
	return out
}

func TestAnswersDisagreementsAreCurrent(t *testing.T) {
	want := wantAnswersDisagreements(t)
	if *updateAnswersDisagreements {
		writeAnswersLines(t, answersDisagreementsFile, answersDisagreementsHeader, want)
	}
	have, exists := readAnswersLines(t, answersDisagreementsFile)
	if !exists {
		t.Fatalf("%s is missing: run with -update-answers-disagreements", answersDisagreementsFile)
	}
	if strings.Join(have, "\n") != strings.Join(want, "\n") {
		t.Errorf("%s is stale: run go test ./internal/cards/effects -run TestAnswersDisagreementsAreCurrent -args -update-answers-disagreements, and list the diff under \"Verdicts changed\" in the PR (ADR 0142 §5)\nhave %d lines, want %d", answersDisagreementsFile, len(have), len(want))
	}
}

// TestAnswersOfKeepsEveryInteractsVerdict is ADR 0142 delivery row 2's
// acceptance test: over every catalog row in scope, answersOf's stack
// tier says what the #2853 abilityInteracts said, frozen below as
// legacyAbilityInteracts. The one intended change is owner answer 2:
// a row whose only answer was a first strike, double strike or
// deathtouch grant is combat tier now. It is deleted in S4, when the
// text read goes.
func TestAnswersOfKeepsEveryInteractsVerdict(t *testing.T) {
	var moved []string
	for _, r := range answersRows(t) {
		if r.mana != nil {
			got, _ := legal.ManaAnswersOf(*r.mana)
			if was := legacySacrificeInteracts(r.mana.SacrificeOther); was != got.HasTier(game.TierStack) {
				t.Errorf("%s: mana interacts was %v, is %v (%s)", r.line(), was, got.HasTier(game.TierStack), got)
			}
			continue
		}
		got, declared := legal.AnswersOf(*r.activated)
		if declared {
			continue // a declaration is reviewed in answers_disagreements.txt
		}
		was, is := legacyAbilityInteracts(*r.activated), got.HasTier(game.TierStack)
		if was == is {
			continue
		}
		if was && !is && got.Has(game.AnswerCombatGrant) && legacyCombatOnly(r.activated.Label) {
			moved = append(moved, r.line())
			continue
		}
		t.Errorf("%s: interacts was %v, is %v (%s)", r.line(), was, is, got)
	}
	sort.Strings(moved)
	t.Logf("%d rows moved to the combat tier (ADR 0142 owner answer 2):\n%s", len(moved), strings.Join(moved, "\n"))
}

// legacyCombatOnly: the #2853 text read matched only on first strike,
// double strike or deathtouch.
func legacyCombatOnly(label string) bool {
	text := strings.ToLower(label)
	if i := strings.Index(text, ":"); i >= 0 {
		text = text[i+1:]
	}
	r := strings.NewReplacer("first strike", "", "double strike", "", "deathtouch", "")
	return legacyEffectTextInteracts(label) && !legacyEffectTextInteracts(": "+r.Replace(text))
}
