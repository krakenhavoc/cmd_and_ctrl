package testsi

import (
	"testing"

	"github.com/google/uuid"

	cardsj "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/j"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestKefkaIsIndestructibleOnlyDuringYourTurn(t *testing.T) {
	g, kefka := effectstest.KefkaGame(t)
	if !effectstest.HasEffectiveKeyword(t, g, kefka, "indestructible") {
		t.Error("Kefka should be indestructible during his controller's turn")
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if effectstest.HasEffectiveKeyword(t, g, kefka, "indestructible") {
		t.Error("Kefka must not be indestructible on an opponent's turn")
	}
}

func TestKefkaExilesOneRandomCardPerOpponentAndTheFreeCastCostsItsOwnerLife(t *testing.T) {
	g, _ := effectstest.KefkaGame(t)
	me, o1, o2 := g.Seats[0], g.Seats[1], g.Seats[2]
	spell := effectstest.GraveyardSpell(o1, "Their Sorcery", "{3}{U}")
	land := uuid.New()
	o2.Graveyard.PushTop(game.Card{InstanceID: land, Name: "Their Island", TypeLine: "Basic Land — Island", Owner: o2.ID, Controller: o2.ID})
	mine := effectstest.GraveyardSpell(me, "My Sorcery", "{2}")

	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Exile.Contains(spell) || !g.Exile.Contains(land) {
		t.Fatal("one card from each opponent's graveyard should be exiled")
	}
	if !me.Graveyard.Contains(mine) {
		t.Error("Kefka's controller's own graveyard is untouched")
	}
	// #4250: the cast is made during the trigger's resolution; a land is
	// not offered.
	offer := effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, spell)
	if effectstest.HasID(offer.DiscardOptions, land) {
		t.Error("a land was offered as a spell")
	}
	before := o1.Life
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("free cast from exile: %v", err)
	}
	effectstest.CastInsideTheResolution(t, g, spell)
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
	g, _ := effectstest.KefkaGame(t)
	o1 := g.Seats[1]
	spell := effectstest.GraveyardSpell(o1, "Their Sorcery", "{3}{U}")
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	// Decline: the resolution_cast's empty answer (#4250), then the turn
	// moves on.
	effectstest.DeclineResolutionCast(t, g, g.Seats[0].ID)
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if !g.Exile.Contains(spell) {
		t.Error("a spell that was never cast stays in exile")
	}
	if o1.Life != 40 {
		t.Errorf("life = %d, want 40: nothing was cast", o1.Life)
	}
}

func TestKefkaCastsOnlyTheExiledCardsNotAnyOtherExiledCard(t *testing.T) {
	g, _ := effectstest.KefkaGame(t)
	me, o1 := g.Seats[0], g.Seats[1]
	effectstest.GraveyardSpell(o1, "Their Sorcery", "{3}{U}")
	other := uuid.New()
	g.Exile.PushTop(game.Card{InstanceID: other, Name: "Old Exiled Spell", TypeLine: "Sorcery", ManaCost: "{1}",
		OracleID: effectstest.SpellControlDivinationOracle, Owner: o1.ID, Controller: o1.ID})
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, other, game.CastSpellParams{FromZone: "exile"}); err == nil {
		t.Error("a card exiled earlier is not covered by Kefka's grant")
	}
}

// "Then each player who owns a spell you cast this way loses life": the
// loss waits for every cast (#4250). With two spells on offer, the first
// cast costs nobody life while the offer is still open; once the second
// is cast, each owner loses its spell's mana value.
func TestKefkaLifeLossWaitsForEveryCast(t *testing.T) {
	g, _ := effectstest.KefkaGame(t)
	me, o1, o2 := g.Seats[0], g.Seats[1], g.Seats[2]
	first := effectstest.GraveyardSpell(o1, "Their Sorcery", "{3}{U}")
	second := effectstest.GraveyardSpell(o2, "Other Sorcery", "{1}{R}")
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, first, second)
	if err := g.CastSpell(me.ID, first, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("first free cast: %v", err)
	}
	if o1.Life != 40 {
		t.Fatalf("life %d after the first cast, want 40: the loss waits for the casts", o1.Life)
	}
	if effectstest.ResolutionCastOfferFor(g, me.ID) == nil {
		t.Fatal("one cast closed the \"any number\" offer")
	}
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("second free cast: %v", err)
	}
	if o1.Life != 36 || o2.Life != 38 {
		t.Errorf("life %d and %d, want 36 and 38 (each spell's mana value)", o1.Life, o2.Life)
	}
	if g.ResolutionPaused() {
		t.Error("the resolution is still held after the last cast")
	}
}

// A table paused on Kefka's offer is a restore point: the restored cast
// runs the life loss once the offer is answered.
func TestKefkaOfferSurvivesARestoreMidPrompt(t *testing.T) {
	g, _ := effectstest.KefkaGame(t)
	o1 := g.Seats[1]
	spell := effectstest.GraveyardSpell(o1, "Their Sorcery", "{3}{U}")
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	r := effectstest.RestoreRoundTrip(t, g, true)
	rme, ro1 := r.Seats[0], r.Seats[1]
	c := effectstest.MustResolutionCastOffer(t, r, rme.ID, game.MayCastKeywordFreeCast, spell)
	if c.PickThen != cardsj.KefkaAfterTheCasts.Key() {
		t.Fatalf("the restored offer's continuation is %q, want %q", c.PickThen, cardsj.KefkaAfterTheCasts.Key())
	}
	if err := effectstest.CastThroughResolutionCast(t, r, rme.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("the restored cast: %v", err)
	}
	if ro1.Life != 36 {
		t.Errorf("life %d on the restored table, want 36", ro1.Life)
	}
}
