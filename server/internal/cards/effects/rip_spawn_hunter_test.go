package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const ripOracle = "c027f369-e445-46a8-94ad-93396321295a"

func ripSeed(p *game.Player, typeLine string, power int) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{InstanceID: id, Name: typeLine, TypeLine: typeLine, Power: power, Toughness: 1, Owner: p.ID, Controller: p.ID})
	return id
}

func TestRipSpawnHunterIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(ripOracle)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Rip: %+v %v", spec, ok)
	}
}

func TestRipSurvivalRevealsXPutsDifferentPowersInHandRestOnBottom(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rip := dskB1Creature(g, me.ID, "Rip, Spawn Hunter", "Legendary Creature — Human Survivor", ripOracle, 4, 4)
	dskB1Tap(g, rip)
	// Pushed top-last: reveal order is a(2), b(2), v(3 Vehicle), land, c(1).
	c := ripSeed(me, "Creature — Bear", 1)
	land := ripSeed(me, "Land", 0)
	v := ripSeed(me, "Artifact — Vehicle", 3)
	b := ripSeed(me, "Creature — Bear", 2)
	a := ripSeed(me, "Creature — Bear", 2)
	advanceTo(t, g, game.StepPostcombatMain)
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick")
	}
	if len(pick.ChooseCards) != 3 {
		t.Fatalf("candidates %d, want the 3 creature/Vehicle cards of the top four (power 4)", len(pick.ChooseCards))
	}
	// The same power twice is refused.
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{a, b}); err == nil {
		t.Fatal("two cards with power 2 were accepted")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{a, v}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(a) || !me.Hand.Contains(v) {
		t.Error("the chosen cards are not in hand")
	}
	for _, id := range []uuid.UUID{b, land} {
		if me.Hand.Contains(id) || !me.Library.Contains(id) {
			t.Error("an unchosen card left the library or reached the hand")
		}
	}
	if !me.Library.Contains(c) {
		t.Error("the fifth card left the library")
	}
}

func TestRipSurvivalDoesNothingUntapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Rip, Spawn Hunter", "Legendary Creature — Human Survivor", ripOracle, false)
	ripSeed(me, "Creature — Bear", 1)
	advanceTo(t, g, game.StepPostcombatMain)
	passPriorityAroundTable(t, g)
	if latestChooseCardsFor(g, me.ID) != nil {
		t.Error("an untapped Rip revealed cards")
	}
}
