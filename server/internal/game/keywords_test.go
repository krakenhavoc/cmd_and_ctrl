package game

import (
	"testing"

	"github.com/google/uuid"
)

// keywords_test.go covers the S18 sub-PR 2 helpers. Each test
// exercises one helper against small card fixtures — the real
// integration path (catalog-declared keywords via the layer engine
// → CardView.Abilities) is covered by sub-PR 4 / 5 catalog-card
// tests.

func cardWithAbilities(abilities ...string) *Card {
	return &Card{
		InstanceID: uuid.New(),
		TypeLine:   "Creature — Test",
		effective: &Characteristic{
			Abilities: append([]string(nil), abilities...),
		},
	}
}

func TestHasKeyword(t *testing.T) {
	c := cardWithAbilities("flying", "vigilance")

	if !HasKeyword(c, "flying") {
		t.Error("expected flying")
	}
	if !HasKeyword(c, "vigilance") {
		t.Error("expected vigilance")
	}
	if HasKeyword(c, "trample") {
		t.Error("did not expect trample")
	}
	if HasKeyword(nil, "flying") {
		t.Error("nil card must return false")
	}
	if HasKeyword(c, "") {
		t.Error("empty kw must return false")
	}
}

func TestHasKeywordOffBattlefieldFallsBackToCatalog(t *testing.T) {
	// Save + restore the hook so this test doesn't leak state.
	prev := CatalogPrintedKeywords
	defer func() { CatalogPrintedKeywords = prev }()

	const flashCardOracle = "ambush-viper-oracle-test"
	CatalogPrintedKeywords = func(oracleID string) []string {
		if oracleID == flashCardOracle {
			return []string{"flash", "deathtouch"}
		}
		return nil
	}

	// Card in hand (no `effective` cache) should read via the
	// catalog fallback.
	c := &Card{
		InstanceID: uuid.New(),
		OracleID:   flashCardOracle,
		TypeLine:   "Creature — Snake",
	}
	if !HasKeyword(c, "flash") {
		t.Error("expected flash via catalog fallback")
	}
	if !HasKeyword(c, "deathtouch") {
		t.Error("expected deathtouch via catalog fallback")
	}
	if HasKeyword(c, "flying") {
		t.Error("did not expect flying via catalog fallback")
	}

	// Card with no OracleID → no catalog lookup → no keywords.
	noOracle := &Card{InstanceID: uuid.New(), TypeLine: "Creature"}
	if HasKeyword(noOracle, "flash") {
		t.Error("card without OracleID must return false")
	}
}

func TestHasKeywordNilHookReturnsFalse(t *testing.T) {
	prev := CatalogPrintedKeywords
	defer func() { CatalogPrintedKeywords = prev }()
	CatalogPrintedKeywords = nil

	c := &Card{InstanceID: uuid.New(), OracleID: "any"}
	if HasKeyword(c, "flash") {
		t.Error("nil hook must return false")
	}
}

func TestHasKeywordPrefersEffectiveOverCatalog(t *testing.T) {
	// A card on the battlefield with a stale/different catalog
	// entry must trust `effective` (which includes both printed
	// keywords via the synthesized static AND grants from other
	// cards' Layer 6 abilities) over the catalog fallback.
	prev := CatalogPrintedKeywords
	defer func() { CatalogPrintedKeywords = prev }()
	CatalogPrintedKeywords = func(oracleID string) []string {
		return []string{"flash"} // catalog says "flash"
	}

	c := &Card{
		InstanceID: uuid.New(),
		OracleID:   "some-oracle",
		effective: &Characteristic{
			Abilities: []string{"flying"}, // effective says "flying"
		},
	}
	if !HasKeyword(c, "flying") {
		t.Error("expected flying from effective")
	}
	if HasKeyword(c, "flash") {
		t.Error("did not expect flash — effective must win over catalog on battlefield")
	}
}

func TestHasSummoningSickness(t *testing.T) {
	// Just arrived, no haste → sick.
	sick := &Card{
		InstanceID:       uuid.New(),
		SummonedThisTurn: true,
		effective:        &Characteristic{Types: []string{"creature"}},
	}
	if !HasSummoningSickness(sick) {
		t.Error("fresh creature without haste should be sick")
	}

	// Just arrived, has haste → not sick.
	hasty := &Card{
		InstanceID:       uuid.New(),
		SummonedThisTurn: true,
		effective: &Characteristic{
			Types:     []string{"creature"},
			Abilities: []string{"haste"},
		},
	}
	if HasSummoningSickness(hasty) {
		t.Error("haste creature must bypass sickness")
	}

	// Been around, no haste → not sick.
	settled := &Card{
		InstanceID: uuid.New(),
		effective:  &Characteristic{Types: []string{"creature"}},
	}
	if HasSummoningSickness(settled) {
		t.Error("creature with SummonedThisTurn=false must not be sick")
	}

	// nil → not sick (avoids defensive nil checks in callers).
	if HasSummoningSickness(nil) {
		t.Error("nil card must not be sick")
	}
}

// TestHasSummoningSicknessIsCreaturesOnly is #530. CR 302.6 gates a
// CREATURE: it can't attack, and can't pay {T} or {Q}, unless its
// controller has controlled it continuously since their most recent
// turn began. A Treasure token, a Sol Ring and a fetchland are not
// creatures and have never been subject to that rule, but this
// helper reported all three as sick because it read
// SummonedThisTurn alone. The engine's write paths papered over
// that with their own `card.IsCreature() &&` prefix; the wire
// (protocol/view.go) did not, so the client greyed abilities the
// server would have allowed. The guard belongs here, where the rule
// is named.
func TestHasSummoningSicknessIsCreaturesOnly(t *testing.T) {
	fresh := func(types ...string) *Card {
		return &Card{
			InstanceID:       uuid.New(),
			SummonedThisTurn: true,
			effective:        &Characteristic{Types: types},
		}
	}

	tests := []struct {
		name string
		card *Card
		want bool
	}{
		{"Treasure token made this turn", fresh("artifact"), false},
		{"fetchland played this turn", fresh("land"), false},
		{"uncrewed Vehicle that entered this turn", fresh("artifact"), false},
		{"enchantment that entered this turn", fresh("enchantment"), false},
		{"planeswalker that entered this turn", fresh("planeswalker"), false},
		{"creature that entered this turn", fresh("creature"), true},
		// An animated land or a crewed Vehicle that ENTERED this
		// turn is a creature whose controller has not controlled it
		// since the turn began, so CR 302.6 still catches it. A
		// Vehicle that has been out since last turn has
		// SummonedThisTurn=false and is judged by the arm above.
		{"artifact creature that entered this turn", fresh("artifact", "creature"), true},
		{"manland animated the turn it entered", fresh("land", "creature"), true},
	}

	for _, tc := range tests {
		if got := HasSummoningSickness(tc.card); got != tc.want {
			t.Errorf("%s: HasSummoningSickness = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// TestCanBlock is the pre-#705 CanBlock(attacker, blocker) table, run
// through its replacement. Every answer is unchanged (ADR 0045
// addendum test plan item 1: pair function parity), and each refusal
// now names its reason.
func TestCanBlock(t *testing.T) {
	g := newActiveGame(t)
	flier := cardWithAbilities("flying")
	vanilla := cardWithAbilities()
	reach := cardWithAbilities("reach")
	doubleFlying := cardWithAbilities("flying")

	for _, tc := range []struct {
		name              string
		attacker, blocker *Card
		want              BlockReason
	}{
		// Flier can be blocked by flier or reach, not by vanilla.
		{"flying blocks flying", flier, doubleFlying, ""},
		{"reach blocks flying", flier, reach, ""},
		{"vanilla can't block flying", flier, vanilla, BlockReasonFlying},
		// Non-flier can be blocked by anything.
		{"vanilla blocks vanilla", vanilla, vanilla, ""},
		{"flier blocks vanilla", vanilla, flier, ""},
		{"reach blocks vanilla", vanilla, reach, ""},
		// Nil guards.
		{"nil attacker", nil, vanilla, BlockReasonCantBeBlocked},
		{"nil blocker", flier, nil, BlockReasonCantBlock},
	} {
		r := g.BlockPairRefusalLocked(tc.attacker, tc.blocker)
		if r.Reason != tc.want {
			t.Errorf("%s: reason = %q, want %q", tc.name, r.Reason, tc.want)
		}
		if got, want := g.CanBlockLocked(tc.attacker, tc.blocker), tc.want == ""; got != want {
			t.Errorf("%s: CanBlockLocked = %v, want %v", tc.name, got, want)
		}
		if (r == BlockOK) != r.Legal() {
			t.Errorf("%s: BlockOK and Legal() disagree for %+v", tc.name, r)
		}
	}
}

func TestBlockerCountValidMenace(t *testing.T) {
	menacer := cardWithAbilities("menace")
	vanilla := cardWithAbilities()
	blk := cardWithAbilities()

	// 0 blockers → always valid (attacker is unblocked).
	if !BlockerCountValid(menacer, nil) {
		t.Error("0 blockers against menace must be valid (unblocked)")
	}

	// 1 blocker against menace → invalid.
	if BlockerCountValid(menacer, []*Card{blk}) {
		t.Error("1 blocker against menace must be invalid")
	}

	// 2 blockers against menace → valid.
	if !BlockerCountValid(menacer, []*Card{blk, blk}) {
		t.Error("2 blockers against menace must be valid")
	}

	// No menace → any count valid.
	if !BlockerCountValid(vanilla, []*Card{blk}) {
		t.Error("1 blocker against non-menace must be valid")
	}

	// Nil attacker → always valid.
	if !BlockerCountValid(nil, []*Card{blk}) {
		t.Error("nil attacker must return true")
	}
}

// TestCanonicalKeyword pins the normalisation the deck importer
// applies to Scryfall's `keywords` array: capitalisation and
// surrounding whitespace are Scryfall's business, the closed set is
// the engine's.
func TestCanonicalKeyword(t *testing.T) {
	for in, want := range map[string]string{
		"Flying":        "flying",
		"First strike":  "first strike",
		"Double strike": "double strike",
		" Flash\n":      "flash",
		"VIGILANCE":     "vigilance",
	} {
		got, ok := CanonicalKeyword(in)
		if !ok || got != want {
			t.Errorf("CanonicalKeyword(%q) = (%q, %v), want (%q, true)", in, got, ok, want)
		}
	}
	// Mechanics the engine does not enforce are dropped rather than
	// passed through as unknown ability strings.
	for _, in := range []string{"Prepared", "Waterbend", "Airbend", "Transform", "Ward", "Cycling", ""} {
		if got, ok := CanonicalKeyword(in); ok {
			t.Errorf("CanonicalKeyword(%q) = (%q, true), want not-canonical", in, got)
		}
	}
}

// TestPrintedCharacteristicMergesKeywordSources covers the overlap
// the #317 / #319 / #320 fix created: a card can now carry printed
// keywords on the Card AND have a catalog entry declaring the same
// ones. Both sources land in Abilities, each keyword once — a
// duplicate is invisible to HasKeyword but renders as a second
// badge on the client's keyword row.
func TestPrintedCharacteristicMergesKeywordSources(t *testing.T) {
	prev := CatalogPrintedKeywords
	defer func() { CatalogPrintedKeywords = prev }()

	const oracle = "ambush-viper-oracle-test"
	CatalogPrintedKeywords = func(oracleID string) []string {
		if oracleID == oracle {
			return []string{"flash", "deathtouch"}
		}
		return nil
	}

	c := Card{
		InstanceID: uuid.New(),
		OracleID:   oracle,
		TypeLine:   "Creature — Snake",
		// The importer stamps the same two off Scryfall, plus one
		// the catalog spec doesn't mention.
		Keywords: []string{"flash", "deathtouch", "reach"},
	}
	got := c.Effective().Abilities
	counts := map[string]int{}
	for _, a := range got {
		counts[a]++
	}
	for _, kw := range []string{"flash", "deathtouch", "reach"} {
		if counts[kw] != 1 {
			t.Errorf("abilities %v: %q appears %d times, want exactly 1", got, kw, counts[kw])
		}
	}
	if len(got) != 3 {
		t.Errorf("abilities = %v, want exactly the three keywords", got)
	}
}
