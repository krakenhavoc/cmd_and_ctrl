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
		effective:        &Characteristic{},
	}
	if !HasSummoningSickness(sick) {
		t.Error("fresh creature without haste should be sick")
	}

	// Just arrived, has haste → not sick.
	hasty := &Card{
		InstanceID:       uuid.New(),
		SummonedThisTurn: true,
		effective:        &Characteristic{Abilities: []string{"haste"}},
	}
	if HasSummoningSickness(hasty) {
		t.Error("haste creature must bypass sickness")
	}

	// Been around, no haste → not sick.
	settled := &Card{
		InstanceID: uuid.New(),
		effective:  &Characteristic{},
	}
	if HasSummoningSickness(settled) {
		t.Error("creature with SummonedThisTurn=false must not be sick")
	}

	// nil → not sick (avoids defensive nil checks in callers).
	if HasSummoningSickness(nil) {
		t.Error("nil card must not be sick")
	}
}

func TestCanBlock(t *testing.T) {
	flier := cardWithAbilities("flying")
	vanilla := cardWithAbilities()
	reach := cardWithAbilities("reach")
	doubleFlying := cardWithAbilities("flying")

	// Flier can be blocked by flier or reach, not by vanilla.
	if !CanBlock(flier, doubleFlying) {
		t.Error("flying should be able to block flying")
	}
	if !CanBlock(flier, reach) {
		t.Error("reach should be able to block flying")
	}
	if CanBlock(flier, vanilla) {
		t.Error("vanilla should NOT be able to block flying")
	}

	// Non-flier can be blocked by anything.
	if !CanBlock(vanilla, vanilla) {
		t.Error("vanilla should block vanilla")
	}
	if !CanBlock(vanilla, flier) {
		t.Error("flier should block vanilla")
	}
	if !CanBlock(vanilla, reach) {
		t.Error("reach should block vanilla")
	}

	// Nil guards.
	if CanBlock(nil, vanilla) {
		t.Error("nil attacker must return false")
	}
	if CanBlock(flier, nil) {
		t.Error("nil blocker must return false")
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
