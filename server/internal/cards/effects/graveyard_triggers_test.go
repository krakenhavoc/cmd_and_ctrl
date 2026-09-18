package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// graveyard_triggers_test.go — the catalog half of #925: the two
// cards whose printed abilities watch from a graveyard. The engine
// half (the zone index, the harvest, the CR 108.4 owner rule, the
// undo across the prompt) is in game/trigger_zones_test.go.

const (
	narcomoebaOracle = "dc65bb62-4ba2-4ec1-b3b9-9c51e64cbfc8"
	bloodghastOracle = "e97f9c2b-b41e-4f36-9245-77c0ac125647"
)

// putOnTopOfLibrary places a catalog card on top of a seat's library
// so a mill can put it into that seat's graveyard from there — the
// exact zone pair Narcomoeba's trigger names.
func putOnTopOfLibrary(g *game.Game, p *game.Player, name, typeLine, oracle string) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   typeLine,
		OracleID:   oracle,
		Owner:      p.ID,
		Controller: p.ID,
	})
	return id
}

func inZone(z *game.Zone, id uuid.UUID) bool {
	if z == nil {
		return false
	}
	for _, c := range z.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

// TestNarcomoebaReturnsItselfWhenMilled: the card is in the graveyard
// when the ability triggers — being put there IS the trigger — so
// this fires only because the harvest walks the graveyard for a card
// that declared it (#925).
func TestNarcomoebaReturnsItselfWhenMilled(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := putOnTopOfLibrary(g, me, "Narcomoeba", "Creature — Illusion", narcomoebaOracle)

	g.WithWriteLock(func() {
		if err := g.MillNForEffect(me.ID, 1); err != nil {
			t.Fatalf("MillNForEffect: %v", err)
		}
	})
	if !inZone(me.Graveyard, id) {
		t.Fatal("the mill did not put Narcomoeba into its owner's graveyard")
	}

	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)

	if inZone(me.Graveyard, id) {
		t.Error("Narcomoeba is still in the graveyard after its trigger resolved")
	}
	if !onBattlefieldByName(g, "Narcomoeba") {
		t.Error("Narcomoeba did not reach the battlefield")
	}
}

// TestNarcomoebaDeclinedStaysInTheGraveyard: the "you may" is a CR
// 603.5 prompt, so "no" drops the trigger and leaves the card where
// it is. It also proves the graveyard trigger takes the ORDINARY
// dispatch rather than a second path of its own.
func TestNarcomoebaDeclinedStaysInTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := putOnTopOfLibrary(g, me, "Narcomoeba", "Creature — Illusion", narcomoebaOracle)

	g.WithWriteLock(func() {
		if err := g.MillNForEffect(me.ID, 1); err != nil {
			t.Fatalf("MillNForEffect: %v", err)
		}
	})
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)

	if !inZone(me.Graveyard, id) {
		t.Error("a declined Narcomoeba left the graveyard anyway")
	}
}

// TestNarcomoebaOnTheBattlefieldDoesNotWatchAMill is the negative
// half, and the assertion that fails with the fix backed out: a
// Narcomoeba already on the battlefield has no graveyard ability,
// and before the Zones dimension the harvester's battlefield walk
// was the ONLY walk — so a card declaring this trigger would have
// fired from play and tried to return itself from a graveyard it is
// not in.
func TestNarcomoebaOnTheBattlefieldDoesNotWatchAMill(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	g.Battlefield.PushTop(game.Card{
		InstanceID: uuid.New(),
		Name:       "Narcomoeba",
		TypeLine:   "Creature — Illusion",
		OracleID:   narcomoebaOracle,
		Power:      1,
		Toughness:  1,
		Owner:      me.ID,
		Controller: me.ID,
	})
	putOnTopOfLibrary(g, me, "Filler", "Creature — Test", "")

	g.WithWriteLock(func() {
		if err := g.MillNForEffect(me.ID, 1); err != nil {
			t.Fatalf("MillNForEffect: %v", err)
		}
	})

	if c := latestTriggerPrompt(g, me.ID); c != nil {
		t.Fatalf("a battlefield Narcomoeba queued a graveyard trigger (%q)", c.Reason)
	}
	if len(g.PendingTriggers) != 0 {
		t.Fatalf("a battlefield Narcomoeba queued %d triggers, want 0", len(g.PendingTriggers))
	}
}

// TestBloodghastReturnsOnLandfallFromTheGraveyard: landfall watched
// from a graveyard, with "you control" and "your graveyard" both
// meaning the card's OWNER (CR 108.4).
func TestBloodghastReturnsOnLandfallFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: id,
		Name:       "Bloodghast",
		TypeLine:   "Creature — Vampire Spirit",
		OracleID:   bloodghastOracle,
		Power:      2,
		Toughness:  1,
		Owner:      me.ID,
	})

	playLandFromHand(t, g, "Swamp", "")

	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)

	if inZone(me.Graveyard, id) {
		t.Error("Bloodghast is still in the graveyard after landfall")
	}
	if !onBattlefieldByName(g, "Bloodghast") {
		t.Error("Bloodghast did not return to the battlefield")
	}
}

// TestBloodghastIgnoresAnOpponentsLand: "a land YOU control", read
// off the graveyard card's OWNER. An opponent's land drop is not
// Bloodghast's landfall.
func TestBloodghastIgnoresAnOpponentsLand(t *testing.T) {
	g := newCatalogGame(t)
	me, opponent := g.Seats[0], g.Seats[1]
	me.Graveyard.PushTop(game.Card{
		InstanceID: uuid.New(),
		Name:       "Bloodghast",
		TypeLine:   "Creature — Vampire Spirit",
		OracleID:   bloodghastOracle,
		Owner:      me.ID,
	})

	// The opponent's land enters directly — no land-drop bookkeeping
	// needed, the trigger watches the ENTRY.
	g.WithWriteLock(func() {
		landID := uuid.New()
		g.Battlefield.PushTop(game.Card{
			InstanceID: landID,
			Name:       "Island",
			TypeLine:   "Basic Land — Island",
			Owner:      opponent.ID,
			Controller: opponent.ID,
		})
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: landID, Actor: opponent.ID})
	})

	if c := latestTriggerPrompt(g, me.ID); c != nil {
		t.Fatalf("an opponent's land woke Bloodghast (%q)", c.Reason)
	}
}

// onBattlefieldByName reports whether a card with that name is on the
// battlefield. Name rather than instance ID because a return from a
// graveyard to the battlefield is a new object (CR 400.7).
func onBattlefieldByName(g *game.Game, name string) bool {
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			return true
		}
	}
	return false
}
