package testsex

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sacrifice another artifact: a +1/+1 counter and a convert, at sorcery
// speed.
func TestFlamewarSacrificesAnArtifactToGrowAndConvert(t *testing.T) {
	g, me, _ := effectstest.FlipMain(t)
	id := effectstest.TransformerIn(t, g, effectstest.FlamewarRow(), me)
	rock := effectstest.PushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)

	effectstest.P7Activate(t, g, me, id, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}})

	c := effectstest.WantFace(t, g, id, 1, "Flamewar, Streetwise Operative")
	if c.Counters["+1/+1"] != 1 {
		t.Errorf("%d +1/+1 counters, want 1", c.Counters["+1/+1"])
	}
	if zone, _ := effectstest.CardWhere(g, me, rock); zone != "graveyard" {
		t.Errorf("the sacrificed artifact is in %s", zone)
	}
}

// Flamewar cannot sacrifice itself ("another artifact").
func TestFlamewarCannotSacrificeItself(t *testing.T) {
	g, me, _ := effectstest.FlipMain(t)
	id := effectstest.TransformerIn(t, g, effectstest.FlamewarRow(), me)
	err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{id}})
	if err == nil {
		t.Fatal("Flamewar sacrificed itself to its own ability")
	}
}

// The back face exiles that many cards face down with intel counters and
// converts back; the front's second ability then takes them.
func TestFlamewarIntelRoundTrip(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	id := effectstest.TransformerConverted(t, g, effectstest.FlamewarRow(), me)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	libBefore := me.Library.Size()

	effectstest.EmitCombatDamage(g, id, opp.ID, 3)
	effectstest.TfSettle(t, g)

	c := effectstest.WantFace(t, g, id, 0, "Flamewar, Brash Veteran")
	if c.TurnOvers != 2 {
		t.Errorf("turn-overs = %d, want 2", c.TurnOvers)
	}
	if me.Library.Size() != libBefore-3 {
		t.Fatalf("library %d → %d, want 3 exiled", libBefore, me.Library.Size())
	}
	intel := 0
	for _, e := range g.Exile.Cards {
		if e.Owner == me.ID && e.Counters["intel"] == 1 {
			intel++
		}
	}
	if intel != 3 {
		t.Fatalf("%d exiled cards carry an intel counter, want 3", intel)
	}

	// An unrelated exiled card of mine has no intel counter and stays.
	plain := uuid.New()
	g.Exile.PushTop(game.Card{InstanceID: plain, Name: "Plain Exile", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	effectstest.HandCardForTest(me, "Chaff", "Sorcery", "")
	effectstest.B06AddMana(me, "C")
	effectstest.P7Activate(t, g, me, id, 1, game.ActivateAbilityParams{})

	if me.Hand.Size() != 3 {
		t.Errorf("hand = %d, want the 3 intel cards (the old hand was the cost)", me.Hand.Size())
	}
	if zone, _ := effectstest.CardWhere(g, me, plain); zone != "exile" {
		t.Errorf("a card without an intel counter went to %s", zone)
	}
	for _, e := range g.Exile.Cards {
		if e.Owner == me.ID && e.Counters["intel"] > 0 {
			t.Errorf("%s is still in exile with an intel counter", e.Name)
		}
	}
}

// Intel cards another player owns are not "exiled cards you own".
func TestFlamewarTakesOnlyCardsYouOwn(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	id := effectstest.TransformerIn(t, g, effectstest.FlamewarRow(), me)
	theirs := uuid.New()
	g.Exile.PushTop(game.Card{InstanceID: theirs, Name: "Their Intel", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID, Counters: map[string]int{"intel": 1}})
	effectstest.B06AddMana(me, "C")
	effectstest.P7Activate(t, g, me, id, 1, game.ActivateAbilityParams{})
	if zone, _ := effectstest.CardWhere(g, opp, theirs); zone != "exile" {
		t.Errorf("an opponent's intel card went to %s", zone)
	}
}
