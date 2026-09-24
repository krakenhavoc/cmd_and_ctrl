package coverage

import (
	"sort"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// oracle_test.go — #1276's gate. See oracle.go for what it checks and
// what it deliberately does not.

// knownOracleMismatches pins every finding the gate tolerates today,
// keyed by OracleFinding.Key, each with a one-line reason. The test
// fails on a set difference in EITHER direction, so the list can only
// shrink: a new finding fails, and a pin that stops matching (the card
// was fixed, renamed or removed) fails too until its row is deleted.
//
// Every MISSING row is on a card that declares CompletenessCaveats and
// whose caveat describes the gap in prose rather than quoting the cost
// (a caveat that quotes the cost, or a loyalty caveat that says "the
// −7", excuses itself and needs no row).
var knownOracleMismatches = map[string]string{
	// Heart of Kiran prints the alternative crew as prose — "You may
	// remove a loyalty counter from a planeswalker you control rather
	// than pay this Vehicle's crew cost" — not as a cost line, so
	// there is nothing for the label's cost to match. The label is
	// also the string the bot's counter-cost test keys on.
	"Heart of Kiran | cost not printed | crew — remove a loyalty counter from a planeswalker you control": "alternative crew cost; the card prints it as prose, not a cost line",

	// No cost shape yet (#1381 lists them).
	"Jarad, Golgari Lich Lord | printed ability not registered | sacrifice a swamp and a forest":                           "two differently-typed sacrifice clauses in one cost",
	"Kozilek, the Great Distortion | printed ability not registered | discard a card with mana value x":                    "X is read off the discarded card",
	"Ruthless Technomancer | printed ability not registered | {2}{b}, sacrifice x artifacts":                               "a sacrifice count of X",
	"Transmutation Font | printed ability not registered | {3}, {t}, sacrifice three artifact tokens with different names": "\"with different names\" has no sacrifice predicate",
}

func loadOracleFixture(t *testing.T) map[string]OracleCard {
	t.Helper()
	fx, err := LoadOracleFixture(OracleFixturePath)
	if err != nil {
		t.Fatalf("%v — regenerate it (oracle_fixture_test.go)", err)
	}
	return fx
}

// TestAbilitiesMatchOracleText is the gate: every registered activated
// and loyalty ability against the printed card, and every printed one
// against the registry.
func TestAbilitiesMatchOracleText(t *testing.T) {
	report := CheckAbilitiesAgainstOracle(effects.All(), loadOracleFixture(t))
	live := map[string]OracleFinding{}
	for _, f := range report.Findings {
		live[f.Key()] = f
	}

	for _, f := range report.Findings {
		if _, ok := knownOracleMismatches[f.Key()]; ok {
			continue
		}
		t.Errorf("%s\n\n%s\n  pin key: %q", f.Describe(), oracleAdvice(f.Kind), f.Key())
	}

	var stale []string
	for key := range knownOracleMismatches {
		if _, ok := live[key]; !ok {
			stale = append(stale, key)
		}
	}
	sort.Strings(stale)
	for _, key := range stale {
		t.Errorf(`knownOracleMismatches pins %q, which no longer applies.

The card was fixed, renamed or removed. Delete the row: a pin that
matches nothing hides the next real mismatch on the same card.`, key)
	}
	t.Logf("%d abilities, %d effects scored, %d printed cost lines; %d findings, %d pinned",
		report.Abilities, report.EffectsCompared, report.PrintedLines, len(report.Findings), len(knownOracleMismatches))
}

func oracleAdvice(k OracleFindingKind) string {
	switch k {
	case OracleNoText:
		return `The fixture has no text for this card. Regenerate it (needs the dump):

  CMDCTRL_SCRYFALL_DUMP=../data/scryfall/default-cards.json \
    go test ./internal/cards/coverage/ -run TestOracleFixtureIsCurrent -update-oracle`
	case OracleAbilityMissing:
		return `The card prints this ability and the Spec registers nothing for it. Register
it, or — if it is deliberately left out — declare CompletenessCaveats with a
caveat that quotes its cost ("The {2}, Discard a card ability isn't
implemented"; for a loyalty ability "The −7 isn't offered"), which excuses
it here. Pin it in knownOracleMismatches only when the caveat can't quote it.`
	default:
		return `Read the card. A label whose cost the card does not print, or whose effect is
not the one printed at that cost, is usually the ability being wrong, not the
label (The Wandering Emperor, #1276). Fix the ability and its Label; pin it in
knownOracleMismatches with a reason only if the label is deliberately not a
printed line.`
	}
}

// TestOracleCheckIsNotVacuous guards the guard. A normaliser or
// splitter that silently stopped matching anything would make the
// gate pass on nothing; these floors are far below today's counts
// (548 / ~450 / ~400) and far above zero.
func TestOracleCheckIsNotVacuous(t *testing.T) {
	r := CheckAbilitiesAgainstOracle(effects.All(), loadOracleFixture(t))
	if r.Abilities < 300 || r.EffectsCompared < 250 || r.PrintedLines < 250 {
		t.Fatalf("the oracle check compared almost nothing: %d abilities, %d effects scored, %d printed cost lines",
			r.Abilities, r.EffectsCompared, r.PrintedLines)
	}
}

// TestOracleCheckCatchesTheWanderingEmperor is the back-out proof,
// kept: the three labels S27 shipped (git show 21f8afd9^, PR #1274)
// run against the real card must fail the EFFECT check three times.
// All three costs are printed on the card, so a cost-only check — the
// one the issue first sketched — would pass them.
func TestOracleCheckCatchesTheWanderingEmperor(t *testing.T) {
	const emperor = "0c7f18d5-36cb-4bc6-a358-443b97666215"
	fx := loadOracleFixture(t)
	if _, ok := fx[emperor]; !ok {
		t.Fatal("fixture has no Wandering Emperor")
	}
	wrong := effects.Spec{
		OracleID: emperor,
		Name:     "The Wandering Emperor",
		Activated: []effects.ActivatedAbility{
			{Label: "+1: Create a 2/2 white Samurai creature token with vigilance.", Cost: effects.LoyaltyCost(1)},
			{Label: "−1: Exile target tapped creature.", Cost: effects.LoyaltyCost(-1)},
			{Label: "−2: Up to one target creature gets +2/+1 and gains lifelink until end of turn.", Cost: effects.LoyaltyCost(-2)},
		},
	}
	got := CheckAbilitiesAgainstOracle([]effects.Spec{wrong}, fx).Findings
	var mismatched []string
	for _, f := range got {
		if f.Kind == OracleEffectMismatch {
			mismatched = append(mismatched, f.Cost)
		}
	}
	sort.Strings(mismatched)
	if strings.Join(mismatched, " ") != "+1 -1 -2" {
		t.Fatalf("want effect mismatches at +1, -1 and -2, got %v\nall findings: %+v", mismatched, got)
	}

	// And the corrected card, as registered today, is clean.
	fixed, ok := effects.Lookup(emperor)
	if !ok {
		t.Fatal("The Wandering Emperor is not registered")
	}
	if f := CheckAbilitiesAgainstOracle([]effects.Spec{fixed}, fx).Findings; len(f) != 0 {
		t.Fatalf("the registered Emperor should be clean, got %+v", f)
	}
}

// oracleCase runs one synthetic Spec against one synthetic card.
func oracleCase(t *testing.T, text string, caveats []string, labels ...string) []OracleFinding {
	t.Helper()
	s := effects.Spec{OracleID: "x", Name: "Grimgrin, Corpse-Born", Caveats: caveats}
	for _, l := range labels {
		s.Activated = append(s.Activated, effects.ActivatedAbility{Label: l})
	}
	fx := map[string]OracleCard{"x": {Name: "Grimgrin, Corpse-Born", Text: text}}
	return CheckAbilitiesAgainstOracle([]effects.Spec{s}, fx).Findings
}

func findingKinds(fs []OracleFinding) string {
	var ks []string
	for _, f := range fs {
		ks = append(ks, string(f.Kind)+" @ "+f.Cost)
	}
	return strings.Join(ks, "; ")
}

// TestOracleNormalisation pins each normalisation the issue named: the
// Unicode minus, reminder text, the card's own name (full and short)
// against "this creature", loyalty brackets, and a station threshold.
func TestOracleNormalisation(t *testing.T) {
	cases := []struct {
		name, text string
		labels     []string
		want       string // findingKinds, "" for clean
	}{
		{"unicode minus vs hyphen", "−2: Destroy target creature.", []string{"-2: Destroy target creature."}, ""},
		{"bracketed loyalty", "0: Draw a card.", []string{"[0]: Draw a card."}, ""},
		{"reminder text on both sides", "Ninjutsu {1}{B} ({1}{B}, Return an unblocked attacker you control to hand: Put this card onto the battlefield.)",
			[]string{"Ninjutsu {1}{B} (reminder)"}, ""},
		{"full name vs this creature", "Sacrifice another creature: Untap this creature and put a +1/+1 counter on it.",
			[]string{"Sacrifice another creature: Untap Grimgrin, Corpse-Born and put a +1/+1 counter on it."}, ""},
		{"short name vs this creature", "{2}, Sacrifice this creature: Draw a card.",
			[]string{"{2}, Sacrifice Grimgrin: Draw a card."}, ""},
		{"station threshold", "Station\n12+ | {3}{W}, {T}: Create a 1/1 white Soldier creature token.",
			[]string{"{3}{W}, {T}: Create a 1/1 white Soldier creature token.", "Station"}, ""},
		{"ability word", "Threshold — {R}, {T}: It deals 2 damage to any target.",
			[]string{"{R}, {T}: It deals 2 damage to any target."}, ""},
		{"keyword boundary: crew 1 is not crew 10", "Crew 10", []string{"Crew 1"}, "cost not printed @ crew 1"},
		{"cost not printed", "{T}: Draw a card.", []string{"{1}, {T}: Draw a card."}, "cost not printed @ {1}, {t}; printed ability not registered @ {t}"},
		{"modal line absorbs its bullets", "{2}, Sacrifice this creature: Choose one —\n• Destroy target artifact.\n• Destroy target enchantment.",
			[]string{"{2}, Sacrifice Grimgrin: Destroy target artifact", "{2}, Sacrifice Grimgrin: Destroy target enchantment"}, ""},
		{"mana ability lines are out of scope", "{T}: Add {G}.\n{2}, {T}: Double the amount of each type of unspent mana you have.", nil, ""},
		{"a colon inside a granted ability is not a line", `Equipped creature has "{T}: Draw a card."`, nil, ""},
		{"missing loyalty ability", "+1: Draw a card.\n−7: You win the game.", []string{"+1: Draw a card."}, "printed ability not registered @ -7"},
		{"effect mismatch at a printed cost", "+1: Draw a card.\n−1: Create a 1/1 white Soldier creature token.",
			[]string{"+1: Create a 1/1 white Soldier creature token.", "−1: Draw a card."}, "effect does not match @ +1; effect does not match @ -1"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := findingKinds(oracleCase(t, tc.text, nil, tc.labels...)); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// TestOracleCaveatExcusesOnlyWhatItNames: a caveat quoting the cost,
// or naming the loyalty number as its own token, excuses a missing
// line; one that merely contains the digits does not.
func TestOracleCaveatExcusesOnlyWhatItNames(t *testing.T) {
	text := "+1: Draw a card.\n−7: Target creature gets -7/-7 until end of turn.\n{2}, Discard a card: Create a Treasure token."
	label := "+1: Draw a card."
	for _, tc := range []struct {
		caveat, want string
	}{
		{"The −7 isn't offered. The {2}, Discard a card ability isn't implemented.", ""},
		{"The ultimate isn't offered.", "printed ability not registered @ -7; printed ability not registered @ {2}, discard a card"},
		{"Creatures get -7/-7 only once.", "printed ability not registered @ -7; printed ability not registered @ {2}, discard a card"},
	} {
		if got := findingKinds(oracleCase(t, text, []string{tc.caveat}, label)); got != tc.want {
			t.Errorf("caveat %q: got %q, want %q", tc.caveat, got, tc.want)
		}
	}
}
