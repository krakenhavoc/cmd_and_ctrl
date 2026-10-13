package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// kuja_a_kefka_test.go — Kefka, Dancing Mad (#3213).

func kefkaGame(t *testing.T) (*game.Game, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	return g, b27Push(g, me.ID, "Kefka, Dancing Mad", "Legendary Creature — Human Wizard", kefkaOracle, "{5}{B}{R}", 6, 6, "B", "R")
}

func graveyardSpell(owner *game.Player, name, cost string) uuid.UUID {
	id := uuid.New()
	owner.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Sorcery", ManaCost: cost,
		OracleID: spellControlDivinationOracle, Owner: owner.ID, Controller: owner.ID})
	return id
}

func TestKefkaIsIndestructibleOnlyDuringYourTurn(t *testing.T) {
	g, kefka := kefkaGame(t)
	if !hasEffectiveKeyword(t, g, kefka, "indestructible") {
		t.Error("Kefka should be indestructible during his controller's turn")
	}
	advanceToNextSeatsTurn(t, g)
	if hasEffectiveKeyword(t, g, kefka, "indestructible") {
		t.Error("Kefka must not be indestructible on an opponent's turn")
	}
}

func TestKefkaExilesOneRandomCardPerOpponentAndTheFreeCastCostsItsOwnerLife(t *testing.T) {
	g, _ := kefkaGame(t)
	me, o1, o2 := g.Seats[0], g.Seats[1], g.Seats[2]
	spell := graveyardSpell(o1, "Their Sorcery", "{3}{U}")
	land := uuid.New()
	o2.Graveyard.PushTop(game.Card{InstanceID: land, Name: "Their Island", TypeLine: "Basic Land — Island", Owner: o2.ID, Controller: o2.ID})
	mine := graveyardSpell(me, "My Sorcery", "{2}")

	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)

	if !g.Exile.Contains(spell) || !g.Exile.Contains(land) {
		t.Fatal("one card from each opponent's graveyard should be exiled")
	}
	if !me.Graveyard.Contains(mine) {
		t.Error("Kefka's controller's own graveyard is untouched")
	}
	before := o1.Life
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("free cast from exile: %v", err)
	}
	if got := o1.Life; got != before-4 {
		t.Errorf("the spell's owner life %d -> %d, want -4 (its mana value)", before, got)
	}
	// A land can't be cast.
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{FromZone: "exile"}); err == nil {
		t.Error("a land exiled this way must not be castable")
	}
	if o2.Life != 40 {
		t.Errorf("the land's owner lost life: %d", o2.Life)
	}
}

func TestKefkaAnUncastCardStaysInExileAndCostsNobodyLife(t *testing.T) {
	g, _ := kefkaGame(t)
	o1 := g.Seats[1]
	spell := graveyardSpell(o1, "Their Sorcery", "{3}{U}")
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	// Decline: let the table pass and the turn move on.
	advanceToNextSeatsTurn(t, g)
	if !g.Exile.Contains(spell) {
		t.Error("a spell that was never cast stays in exile")
	}
	if o1.Life != 40 {
		t.Errorf("life = %d, want 40: nothing was cast", o1.Life)
	}
}

func TestKefkaCastsOnlyTheExiledCardsNotAnyOtherExiledCard(t *testing.T) {
	g, _ := kefkaGame(t)
	me, o1 := g.Seats[0], g.Seats[1]
	graveyardSpell(o1, "Their Sorcery", "{3}{U}")
	other := uuid.New()
	g.Exile.PushTop(game.Card{InstanceID: other, Name: "Old Exiled Spell", TypeLine: "Sorcery", ManaCost: "{1}",
		OracleID: spellControlDivinationOracle, Owner: o1.ID, Controller: o1.ID})
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, other, game.CastSpellParams{FromZone: "exile"}); err == nil {
		t.Error("a card exiled earlier is not covered by Kefka's grant")
	}
}
