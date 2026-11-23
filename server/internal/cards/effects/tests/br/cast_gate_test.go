package testsbr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestRuleOfLawRefusesTheSecondSpell is the CastTally shape, and it
// also pins that the FIRST spell is fine — a gate that refused
// everything would pass a one-sided test.
func TestRuleOfLawRefusesTheSecondSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedEnchantment(g, effectstest.RuleOfLawOracle, "Rule of Law", "Enchantment", active.ID)

	// First spell of the turn: allowed.
	if _, err := effectstest.CastWithOptionalCosts(t, g, "Burst Lightning", "Instant",
		effectstest.BurstLightningOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}}, nil, nil); err != nil {
		t.Fatalf("first spell under Rule of Law was refused: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	// Second: refused at announce.
	id, err := effectstest.CastWithOptionalCosts(t, g, "Burst Lightning", "Instant",
		effectstest.BurstLightningOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}}, nil, nil)
	effectstest.AssertCantCast(t, err, "more than one spell")
	if !active.Hand.Contains(id) {
		t.Errorf("a refused cast did not leave the card in hand")
	}
}

// TestRuleOfLawFromAnOpponentsBattlefield is the "a restriction from
// an OPPONENT's permanent" assertion: Rule of Law is symmetrical and
// nothing about the gate is keyed to who controls the source.
func TestRuleOfLawFromAnOpponentsBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opponent := g.Seats[1]
	if active.ID == opponent.ID {
		t.Fatalf("seat setup: active and opponent are the same seat")
	}
	effectstest.SeedEnchantment(g, effectstest.RuleOfLawOracle, "Rule of Law", "Enchantment", opponent.ID)

	if _, err := effectstest.CastWithOptionalCosts(t, g, "Burst Lightning", "Instant",
		effectstest.BurstLightningOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opponent.ID}}, nil, nil); err != nil {
		t.Fatalf("first spell was refused: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	_, err := effectstest.CastWithOptionalCosts(t, g, "Burst Lightning", "Instant",
		effectstest.BurstLightningOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opponent.ID}}, nil, nil)
	effectstest.AssertCantCast(t, err, "more than one spell")
}

// TestRuleOfLawStopsRestrictingWhenItLeaves pins that nothing is
// stored: the restriction is derived from the battlefield, so
// removing the source lifts it on the very next query.
func TestRuleOfLawStopsRestrictingWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	law := effectstest.SeedEnchantment(g, effectstest.RuleOfLawOracle, "Rule of Law", "Enchantment", active.ID)

	if _, err := effectstest.CastWithOptionalCosts(t, g, "Burst Lightning", "Instant",
		effectstest.BurstLightningOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}}, nil, nil); err != nil {
		t.Fatalf("first spell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	g.WithWriteLock(func() {
		_ = g.ExileCardForEffect(law)
	})
	if _, err := effectstest.CastWithOptionalCosts(t, g, "Burst Lightning", "Instant",
		effectstest.BurstLightningOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}}, nil, nil); err != nil {
		t.Errorf("the restriction outlived its source: %v", err)
	}
}

// TestGrafdiggersCageRefusesAGraveyardCast is the ZONE shape, and the
// one that proves the gate sees where a cast comes from.
func TestGrafdiggersCageRefusesAGraveyardCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedEnchantment(g, effectstest.GrafdiggersCageOracle, "Grafdigger's Cage", "Artifact", active.ID)

	// Faithless Looting prints flashback, so its graveyard cast is a
	// legal one for the Cage to refuse — the announce gets all the way
	// past the cast-path and alternative-cost checks and then hits the
	// gate. That is why the assertion is the SENTINEL and not merely
	// "an error": a card the catalog does not open for graveyard casts
	// is refused earlier, by ErrCastZoneNotAllowed, and would make
	// this test vacuous.
	id, err := effectstest.CastFlashbackFromGraveyard(t, g, effectstest.FaithlessLootingOracle)
	effectstest.AssertCantCast(t, err, "graveyards or libraries")
	if !active.Graveyard.Contains(id) {
		t.Errorf("a refused flashback did not leave the card in the graveyard")
	}

	// Without the Cage the same announce is accepted, so the refusal
	// above is the Cage's and not the harness's.
	g2 := effectstest.NewCatalogGame(t)
	if _, err := effectstest.CastFlashbackFromGraveyard(t, g2, effectstest.FaithlessLootingOracle); err != nil {
		t.Fatalf("flashback with no Cage out was refused: %v", err)
	}
}

// TestUrzasRuinousBlastNeedsALegendary is CR 205.4e — the spell's own
// condition, the gate's other source.
func TestUrzasRuinousBlastNeedsALegendary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]

	id, err := effectstest.CastWithOptionalCosts(t, g, "Urza's Ruinous Blast",
		"Legendary Sorcery", effectstest.UrzasRuinousOracle, nil, nil, nil)
	effectstest.AssertCantCast(t, err, "legendary creature or planeswalker")
	if !active.Hand.Contains(id) {
		t.Errorf("a refused legendary sorcery did not leave the card in hand")
	}

	// With a legendary creature out it casts.
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Legendary Bear",
		TypeLine:   "Legendary Creature — Bear",
		Power:      2,
		Toughness:  2,
		Owner:      active.ID,
		Controller: active.ID,
	})
	if _, err := effectstest.CastWithOptionalCosts(t, g, "Urza's Ruinous Blast",
		"Legendary Sorcery", effectstest.UrzasRuinousOracle, nil, nil, nil); err != nil {
		t.Errorf("Urza's Ruinous Blast was refused with a legendary creature out: %v", err)
	}
}

// TestRakdosLordOfRiotsNeedsAnOpponentToHaveLostLife is the third
// shape: a condition on the card that reads a per-turn tally rather
// than the board.
func TestRakdosLordOfRiotsNeedsAnOpponentToHaveLostLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	victim := g.Seats[1]

	_, err := effectstest.CastWithOptionalCosts(t, g, "Rakdos, Lord of Riots",
		"Legendary Creature — Demon", effectstest.RakdosLordOracle, nil, nil, nil)
	effectstest.AssertCantCast(t, err, "lost life this turn")

	// Bolt the opponent, and Rakdos is castable.
	if _, err := effectstest.CastWithOptionalCosts(t, g, "Burst Lightning", "Instant",
		effectstest.BurstLightningOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}}, nil, nil); err != nil {
		t.Fatalf("Burst Lightning: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if _, err := effectstest.CastWithOptionalCosts(t, g, "Rakdos, Lord of Riots",
		"Legendary Creature — Demon", effectstest.RakdosLordOracle, nil, nil, nil); err != nil {
		t.Errorf("Rakdos was refused after an opponent lost life: %v", err)
	}
}
