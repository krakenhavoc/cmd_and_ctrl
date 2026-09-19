package effects

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cast_gate_test.go — ADR 0073 §7 (#760), the card half: each of the
// four printed shapes refused at ANNOUNCE, with the printed clause in
// the error, and the card left where it was.
//
// The three-way agreement (engine / enumerator / view) is asserted in
// internal/legal and internal/protocol, against the same cards.

// seedEnchantment puts a catalog permanent on the battlefield so its
// CastRestrictions become active — the restriction is derived from
// the battlefield on every query, so this is the whole setup.
func seedEnchantment(g *game.Game, oracleID, name, typeLine string, controller uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       name,
		TypeLine:   typeLine,
		OracleID:   oracleID,
		Owner:      controller,
		Controller: controller,
	})
}

// assertCantCast pins the error SHAPE as well as the sentinel: the
// client renders the printed clause, so a refusal with no reason is a
// toast that says nothing.
func assertCantCast(t *testing.T, err error, wantSubstring string) {
	t.Helper()
	if err == nil {
		t.Fatalf("cast was allowed, want a refusal mentioning %q", wantSubstring)
	}
	if !errors.Is(err, game.ErrCantCast) {
		t.Fatalf("refusal is %v, want it to wrap game.ErrCantCast", err)
	}
	var cant *game.CantCastError
	if !errors.As(err, &cant) {
		t.Fatalf("refusal is %v, want a *game.CantCastError carrying the printed clause", err)
	}
	if cant.Reason == "" {
		t.Fatalf("refusal carries no printed clause")
	}
	if !strings.Contains(cant.Reason, wantSubstring) {
		t.Errorf("refusal reason %q does not mention %q", cant.Reason, wantSubstring)
	}
}

// TestRuleOfLawRefusesTheSecondSpell is the CastTally shape, and it
// also pins that the FIRST spell is fine — a gate that refused
// everything would pass a one-sided test.
func TestRuleOfLawRefusesTheSecondSpell(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	seedEnchantment(g, ruleOfLawOracle, "Rule of Law", "Enchantment", active.ID)

	// First spell of the turn: allowed.
	if _, err := castWithOptionalCosts(t, g, "Burst Lightning", "Instant",
		burstLightningOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}}, nil, nil); err != nil {
		t.Fatalf("first spell under Rule of Law was refused: %v", err)
	}
	passPriorityAroundTable(t, g)

	// Second: refused at announce.
	id, err := castWithOptionalCosts(t, g, "Burst Lightning", "Instant",
		burstLightningOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}}, nil, nil)
	assertCantCast(t, err, "more than one spell")
	if !active.Hand.Contains(id) {
		t.Errorf("a refused cast did not leave the card in hand")
	}
}

// TestRuleOfLawFromAnOpponentsBattlefield is the "a restriction from
// an OPPONENT's permanent" assertion: Rule of Law is symmetrical and
// nothing about the gate is keyed to who controls the source.
func TestRuleOfLawFromAnOpponentsBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opponent := g.Seats[1]
	if active.ID == opponent.ID {
		t.Fatalf("seat setup: active and opponent are the same seat")
	}
	seedEnchantment(g, ruleOfLawOracle, "Rule of Law", "Enchantment", opponent.ID)

	if _, err := castWithOptionalCosts(t, g, "Burst Lightning", "Instant",
		burstLightningOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opponent.ID}}, nil, nil); err != nil {
		t.Fatalf("first spell was refused: %v", err)
	}
	passPriorityAroundTable(t, g)

	_, err := castWithOptionalCosts(t, g, "Burst Lightning", "Instant",
		burstLightningOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opponent.ID}}, nil, nil)
	assertCantCast(t, err, "more than one spell")
}

// TestRuleOfLawStopsRestrictingWhenItLeaves pins that nothing is
// stored: the restriction is derived from the battlefield, so
// removing the source lifts it on the very next query.
func TestRuleOfLawStopsRestrictingWhenItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	law := seedEnchantment(g, ruleOfLawOracle, "Rule of Law", "Enchantment", active.ID)

	if _, err := castWithOptionalCosts(t, g, "Burst Lightning", "Instant",
		burstLightningOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}}, nil, nil); err != nil {
		t.Fatalf("first spell: %v", err)
	}
	passPriorityAroundTable(t, g)

	g.WithWriteLock(func() {
		_ = g.ExileCardForEffect(law)
	})
	if _, err := castWithOptionalCosts(t, g, "Burst Lightning", "Instant",
		burstLightningOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}}, nil, nil); err != nil {
		t.Errorf("the restriction outlived its source: %v", err)
	}
}

// TestGrafdiggersCageRefusesAGraveyardCast is the ZONE shape, and the
// one that proves the gate sees where a cast comes from.
func TestGrafdiggersCageRefusesAGraveyardCast(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	seedEnchantment(g, grafdiggersCageOracle, "Grafdigger's Cage", "Artifact", active.ID)

	// Faithless Looting prints flashback, so its graveyard cast is a
	// legal one for the Cage to refuse — the announce gets all the way
	// past the cast-path and alternative-cost checks and then hits the
	// gate. That is why the assertion is the SENTINEL and not merely
	// "an error": a card the catalog does not open for graveyard casts
	// is refused earlier, by ErrCastZoneNotAllowed, and would make
	// this test vacuous.
	id, err := castFlashbackFromGraveyard(t, g, faithlessLootingOracle)
	assertCantCast(t, err, "graveyards or libraries")
	if !active.Graveyard.Contains(id) {
		t.Errorf("a refused flashback did not leave the card in the graveyard")
	}

	// Without the Cage the same announce is accepted, so the refusal
	// above is the Cage's and not the harness's.
	g2 := newCatalogGame(t)
	if _, err := castFlashbackFromGraveyard(t, g2, faithlessLootingOracle); err != nil {
		t.Fatalf("flashback with no Cage out was refused: %v", err)
	}
}

// castFlashbackFromGraveyard announces a flashback cast out of the
// active seat's own graveyard, claiming the printed offer.
func castFlashbackFromGraveyard(t *testing.T, g *game.Game, oracleID string) (uuid.UUID, error) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Graveyard.PushTop(game.Card{
		InstanceID: id,
		Name:       "Faithless Looting",
		TypeLine:   "Sorcery",
		OracleID:   oracleID,
		Owner:      active.ID,
		Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	return id, g.CastSpell(active.ID, id, game.CastSpellParams{
		FromZone:        "graveyard",
		AlternativeCost: "flashback",
	})
}

// TestUrzasRuinousBlastNeedsALegendary is CR 307.6 — the spell's own
// condition, the gate's other source.
func TestUrzasRuinousBlastNeedsALegendary(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]

	id, err := castWithOptionalCosts(t, g, "Urza's Ruinous Blast",
		"Legendary Sorcery", urzasRuinousOracle, nil, nil, nil)
	assertCantCast(t, err, "legendary creature or planeswalker")
	if !active.Hand.Contains(id) {
		t.Errorf("a refused legendary sorcery did not leave the card in hand")
	}

	// With a legendary creature out it casts.
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Legendary Bear",
		TypeLine:   "Legendary Creature — Bear",
		Power:      2,
		Toughness:  2,
		Owner:      active.ID,
		Controller: active.ID,
	})
	if _, err := castWithOptionalCosts(t, g, "Urza's Ruinous Blast",
		"Legendary Sorcery", urzasRuinousOracle, nil, nil, nil); err != nil {
		t.Errorf("Urza's Ruinous Blast was refused with a legendary creature out: %v", err)
	}
}

// TestRakdosLordOfRiotsNeedsAnOpponentToHaveLostLife is the third
// shape: a condition on the card that reads a per-turn tally rather
// than the board.
func TestRakdosLordOfRiotsNeedsAnOpponentToHaveLostLife(t *testing.T) {
	g := newCatalogGame(t)
	victim := g.Seats[1]

	_, err := castWithOptionalCosts(t, g, "Rakdos, Lord of Riots",
		"Legendary Creature — Demon", rakdosLordOracle, nil, nil, nil)
	assertCantCast(t, err, "lost life this turn")

	// Bolt the opponent, and Rakdos is castable.
	if _, err := castWithOptionalCosts(t, g, "Burst Lightning", "Instant",
		burstLightningOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}}, nil, nil); err != nil {
		t.Fatalf("Burst Lightning: %v", err)
	}
	passPriorityAroundTable(t, g)

	if _, err := castWithOptionalCosts(t, g, "Rakdos, Lord of Riots",
		"Legendary Creature — Demon", rakdosLordOracle, nil, nil, nil); err != nil {
		t.Errorf("Rakdos was refused after an opponent lost life: %v", err)
	}
}
