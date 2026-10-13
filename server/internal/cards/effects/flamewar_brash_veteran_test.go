package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func flamewarRow() cards.Card {
	return transformerRow(flamewarOracleID,
		disturbFace{"Flamewar, Brash Veteran", "Legendary Artifact Creature — Robot", "{1}{B}{R}", "3", "2"},
		disturbFace{"Flamewar, Streetwise Operative", "Legendary Artifact — Vehicle", "", "2", "1"},
		"More Than Meets the Eye {B}{R}\nSacrifice another artifact: Put a +1/+1 counter on Flamewar and convert it. Activate only as a sorcery.\n{1}, Discard your hand: Put all exiled cards you own with intel counters on them into your hand.",
		"Living metal\nMenace, deathtouch\nWhenever Flamewar deals combat damage to a player, exile that many cards from the top of your library face down. Put an intel counter on each of them. Convert Flamewar.",
		"Living metal", "Convert", "Menace", "Deathtouch")
}

// Sacrifice another artifact: a +1/+1 counter and a convert, at sorcery
// speed.
func TestFlamewarSacrificesAnArtifactToGrowAndConvert(t *testing.T) {
	g, me, _ := flipMain(t)
	id := transformerIn(t, g, flamewarRow(), me)
	rock := pushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)

	p7Activate(t, g, me, id, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}})

	c := wantFace(t, g, id, 1, "Flamewar, Streetwise Operative")
	if c.Counters["+1/+1"] != 1 {
		t.Errorf("%d +1/+1 counters, want 1", c.Counters["+1/+1"])
	}
	if zone, _ := cardWhere(g, me, rock); zone != "graveyard" {
		t.Errorf("the sacrificed artifact is in %s", zone)
	}
}

// Flamewar cannot sacrifice itself ("another artifact").
func TestFlamewarCannotSacrificeItself(t *testing.T) {
	g, me, _ := flipMain(t)
	id := transformerIn(t, g, flamewarRow(), me)
	err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{id}})
	if err == nil {
		t.Fatal("Flamewar sacrificed itself to its own ability")
	}
}

// The back face exiles that many cards face down with intel counters and
// converts back; the front's second ability then takes them.
func TestFlamewarIntelRoundTrip(t *testing.T) {
	g, me, opp := flipMain(t)
	id := transformerConverted(t, g, flamewarRow(), me)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	libBefore := me.Library.Size()

	emitCombatDamage(g, id, opp.ID, 3)
	tfSettle(t, g)

	c := wantFace(t, g, id, 0, "Flamewar, Brash Veteran")
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
	handCardForTest(me, "Chaff", "Sorcery", "")
	b06AddMana(me, "C")
	p7Activate(t, g, me, id, 1, game.ActivateAbilityParams{})

	if me.Hand.Size() != 3 {
		t.Errorf("hand = %d, want the 3 intel cards (the old hand was the cost)", me.Hand.Size())
	}
	if zone, _ := cardWhere(g, me, plain); zone != "exile" {
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
	g, me, opp := flipMain(t)
	id := transformerIn(t, g, flamewarRow(), me)
	theirs := uuid.New()
	g.Exile.PushTop(game.Card{InstanceID: theirs, Name: "Their Intel", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID, Counters: map[string]int{"intel": 1}})
	b06AddMana(me, "C")
	p7Activate(t, g, me, id, 1, game.ActivateAbilityParams{})
	if zone, _ := cardWhere(g, opp, theirs); zone != "exile" {
		t.Errorf("an opponent's intel card went to %s", zone)
	}
}
