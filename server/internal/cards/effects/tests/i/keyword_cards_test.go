package testsi

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestSerraAngelHasFlyingVigilance(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Serra Angel",
		TypeLine:   "Creature — Angel",
		OracleID:   effectstest.SerraAngelOracle,
		Power:      4,
		Toughness:  4,
		Owner:      g.Seats[0].ID,
		Controller: g.Seats[0].ID,
	})
	effectstest.AssertKeywords(t, g, id, "flying", "vigilance")
}

func TestColossalDreadmawHasTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Colossal Dreadmaw",
		TypeLine:   "Creature — Dinosaur",
		OracleID:   effectstest.ColossalDreadmawOracle,
		Power:      6,
		Toughness:  6,
		Owner:      g.Seats[0].ID,
		Controller: g.Seats[0].ID,
	})
	effectstest.AssertKeywords(t, g, id, "trample")
}

func TestGiantSpiderHasReach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Giant Spider",
		TypeLine:   "Creature — Spider",
		OracleID:   effectstest.GiantSpiderOracle,
		Power:      2,
		Toughness:  4,
		Owner:      g.Seats[0].ID,
		Controller: g.Seats[0].ID,
	})
	effectstest.AssertKeywords(t, g, id, "reach")
}

func TestTyphoidRatsHasDeathtouch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Typhoid Rats",
		TypeLine:   "Creature — Rat",
		OracleID:   effectstest.TyphoidRatsOracle,
		Power:      1,
		Toughness:  1,
		Owner:      g.Seats[0].ID,
		Controller: g.Seats[0].ID,
	})
	effectstest.AssertKeywords(t, g, id, "deathtouch")
}

func TestYouthfulKnightHasFirstStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Youthful Knight",
		TypeLine:   "Creature — Human Knight",
		OracleID:   effectstest.YouthfulKnightOracle,
		Power:      2,
		Toughness:  1,
		Owner:      g.Seats[0].ID,
		Controller: g.Seats[0].ID,
	})
	effectstest.AssertKeywords(t, g, id, "first strike")
}

func TestFencingAceHasDoubleStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Fencing Ace",
		TypeLine:   "Creature — Human Soldier",
		OracleID:   effectstest.FencingAceOracle,
		Power:      1,
		Toughness:  1,
		Owner:      g.Seats[0].ID,
		Controller: g.Seats[0].ID,
	})
	effectstest.AssertKeywords(t, g, id, "double strike")
}

func TestLightningElementalHasHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Lightning Elemental",
		TypeLine:   "Creature — Elemental",
		OracleID:   effectstest.LightningElementalOracle,
		Power:      4,
		Toughness:  1,
		Owner:      g.Seats[0].ID,
		Controller: g.Seats[0].ID,
	})
	effectstest.AssertKeywords(t, g, id, "haste")
}

func TestWallOfStoneHasDefender(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Wall of Stone",
		TypeLine:   "Creature — Wall",
		OracleID:   effectstest.WallOfStoneOracle,
		Power:      0,
		Toughness:  8,
		Owner:      g.Seats[0].ID,
		Controller: g.Seats[0].ID,
	})
	effectstest.AssertKeywords(t, g, id, "defender")
}

// Off-battlefield fallback: a catalog card with PrintedKeywords
// surfaces those keywords via CatalogPrintedKeywords even when
// the card has no `effective` cache (e.g. a hand-resident Ambush
// Viper needs this for flash gating).
func TestCatalogPrintedKeywordsFallbackOffBattlefield(t *testing.T) {
	// No battlefield push — construct a hand-resident card.
	c := &game.Card{
		InstanceID: uuid.New(),
		OracleID:   effectstest.SerraAngelOracle,
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
