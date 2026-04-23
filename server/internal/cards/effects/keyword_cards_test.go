package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// keyword_cards_test.go covers the S18 sub-PR 4 vanilla keyword
// card batch. Each test pushes the catalog card onto the
// battlefield, forces a layer recompute via ReadSnapshot, and
// asserts the expected keyword strings land in
// Effective().Abilities. The actual combat behaviour (flying block
// restriction, trample overflow, etc.) is tested in
// server/internal/game/combat_test.go against manufactured
// battlefield state; these tests just prove the card files declare
// the right PrintedKeywords slot.

const (
	serraAngelOracle         = "4b7ac066-e5c7-43e6-9e7e-2739b24a905d"
	colossalDreadmawOracle   = "08c7db90-c0cf-4482-b7ee-bb033e5996d2"
	giantSpiderOracle        = "e740ce2f-2134-473c-afa1-1b6d2d1e38ef"
	typhoidRatsOracle        = "d6ee6cc1-902d-4f56-afa5-6fa4813bfbbc"
	youthfulKnightOracle     = "ef2a24f5-ce5e-4054-843a-2cae0c66318a"
	fencingAceOracle         = "2f810936-2ba6-4c2b-84a0-ff4c1deb026b"
	lightningElementalOracle = "58aee5cb-7b88-446e-ab10-9f83c10d7227"
	wallOfStoneOracle        = "cd4cadb4-3156-49bd-b36e-12ba5c85938b"
)

func assertKeywords(t *testing.T, g *game.Game, cardID uuid.UUID, want ...string) {
	t.Helper()
	got := effectiveAbilities(t, g, cardID)
	for _, kw := range want {
		if !containsString(got, kw) {
			t.Errorf("card %s missing %q in effective abilities %v", cardID, kw, got)
		}
	}
}

func TestSerraAngelHasFlyingVigilance(t *testing.T) {
	g := newCatalogGame(t)
	id := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Serra Angel",
		TypeLine:   "Creature — Angel",
		OracleID:   serraAngelOracle,
		Power:      4,
		Toughness:  4,
		Owner:      g.Seats[0].ID,
		Controller: g.Seats[0].ID,
	})
	assertKeywords(t, g, id, "flying", "vigilance")
}

func TestColossalDreadmawHasTrample(t *testing.T) {
	g := newCatalogGame(t)
	id := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Colossal Dreadmaw",
		TypeLine:   "Creature — Dinosaur",
		OracleID:   colossalDreadmawOracle,
		Power:      6,
		Toughness:  6,
		Owner:      g.Seats[0].ID,
		Controller: g.Seats[0].ID,
	})
	assertKeywords(t, g, id, "trample")
}

func TestGiantSpiderHasReach(t *testing.T) {
	g := newCatalogGame(t)
	id := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Giant Spider",
		TypeLine:   "Creature — Spider",
		OracleID:   giantSpiderOracle,
		Power:      2,
		Toughness:  4,
		Owner:      g.Seats[0].ID,
		Controller: g.Seats[0].ID,
	})
	assertKeywords(t, g, id, "reach")
}

func TestTyphoidRatsHasDeathtouch(t *testing.T) {
	g := newCatalogGame(t)
	id := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Typhoid Rats",
		TypeLine:   "Creature — Rat",
		OracleID:   typhoidRatsOracle,
		Power:      1,
		Toughness:  1,
		Owner:      g.Seats[0].ID,
		Controller: g.Seats[0].ID,
	})
	assertKeywords(t, g, id, "deathtouch")
}

func TestYouthfulKnightHasFirstStrike(t *testing.T) {
	g := newCatalogGame(t)
	id := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Youthful Knight",
		TypeLine:   "Creature — Human Knight",
		OracleID:   youthfulKnightOracle,
		Power:      2,
		Toughness:  1,
		Owner:      g.Seats[0].ID,
		Controller: g.Seats[0].ID,
	})
	assertKeywords(t, g, id, "first strike")
}

func TestFencingAceHasDoubleStrike(t *testing.T) {
	g := newCatalogGame(t)
	id := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Fencing Ace",
		TypeLine:   "Creature — Human Soldier",
		OracleID:   fencingAceOracle,
		Power:      1,
		Toughness:  1,
		Owner:      g.Seats[0].ID,
		Controller: g.Seats[0].ID,
	})
	assertKeywords(t, g, id, "double strike")
}

func TestLightningElementalHasHaste(t *testing.T) {
	g := newCatalogGame(t)
	id := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Lightning Elemental",
		TypeLine:   "Creature — Elemental",
		OracleID:   lightningElementalOracle,
		Power:      4,
		Toughness:  1,
		Owner:      g.Seats[0].ID,
		Controller: g.Seats[0].ID,
	})
	assertKeywords(t, g, id, "haste")
}

func TestWallOfStoneHasDefender(t *testing.T) {
	g := newCatalogGame(t)
	id := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Wall of Stone",
		TypeLine:   "Creature — Wall",
		OracleID:   wallOfStoneOracle,
		Power:      0,
		Toughness:  8,
		Owner:      g.Seats[0].ID,
		Controller: g.Seats[0].ID,
	})
	assertKeywords(t, g, id, "defender")
}

// Off-battlefield fallback: a catalog card with PrintedKeywords
// surfaces those keywords via CatalogPrintedKeywords even when
// the card has no `effective` cache (e.g. a hand-resident Ambush
// Viper needs this for flash gating).
func TestCatalogPrintedKeywordsFallbackOffBattlefield(t *testing.T) {
	// No battlefield push — construct a hand-resident card.
	c := &game.Card{
		InstanceID: uuid.New(),
		OracleID:   serraAngelOracle,
		TypeLine:   "Creature — Angel",
	}
	if !game.HasKeyword(c, "flying") {
		t.Errorf("hand-resident Serra Angel: expected flying via catalog fallback")
	}
	if !game.HasKeyword(c, "vigilance") {
		t.Errorf("hand-resident Serra Angel: expected vigilance via catalog fallback")
	}
	if game.HasKeyword(c, "trample") {
		t.Errorf("hand-resident Serra Angel: did not expect trample")
	}
}
