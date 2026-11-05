package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	krangOracle             = "466d5226-f4c7-4d69-9f56-4f893010127f"
	krangChasmSkulkerOracle = "69facbc7-3859-4716-b627-5199571fb3cf"
)

// krangTable seeds Krang for seat 1 at seat 0's main phase. Seat 0 has
// drawn its turn-1 card; seats 1-3 have drawn nothing this turn.
func krangTable(t *testing.T) (g *game.Game, me *game.Player, krang uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	advanceToMain(t, g)
	me = g.Seats[1]
	krang = pushCatalogPermanent(g, me.ID, "Krang, the All-Powerful", "Legendary Artifact Creature — Utrom Robot", krangOracle, false)
	battlefieldCardPtr(t, g, krang).Power, battlefieldCardPtr(t, g, krang).Toughness = 3, 3
	return g, me, krang
}

func krangDraw(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	g.WithWriteLock(func() { _ = g.DrawNForEffect(p.ID, 1) })
	mscM12Settle(t, g)
}

func krangCounters(g *game.Game, id uuid.UUID) int {
	return findBattlefieldCardForTest(g, id).Counters[game.CounterPlusOne]
}

func TestKrangTheAllPowerfulIsFull(t *testing.T) {
	spec, ok := Lookup(krangOracle)
	if !ok {
		t.Fatal("Krang is not registered")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("completeness %s, caveats %v", spec.Completeness, spec.Caveats)
	}
	if len(spec.TriggerDoublers) != 1 || len(spec.Triggered) != 1 {
		t.Errorf("%d doublers, %d triggers, want 1 and 1", len(spec.TriggerDoublers), len(spec.Triggered))
	}
}

// Krang's own trigger is caused by a player drawing a card, so Krang
// doubles it: two counters on each player's second card, none on any
// other card.
func TestKrangTheAllPowerfulSecondCardTriggersTwiceForEachPlayer(t *testing.T) {
	g, me, krang := krangTable(t)
	krangDraw(t, g, me)
	if got := krangCounters(g, krang); got != 0 {
		t.Fatalf("after my first draw: %d counters, want 0", got)
	}
	krangDraw(t, g, me)
	if got := krangCounters(g, krang); got != 2 {
		t.Fatalf("after my second draw: %d counters, want 2 (the trigger and its additional instance)", got)
	}
	krangDraw(t, g, me)
	if got := krangCounters(g, krang); got != 2 {
		t.Fatalf("after my third draw: %d counters, want still 2", got)
	}
	// Seat 0 drew its turn-1 card already: this is its second.
	krangDraw(t, g, g.Seats[0])
	if got := krangCounters(g, krang); got != 4 {
		t.Fatalf("after an opponent's second draw: %d counters, want 4", got)
	}
	if got := hobA4Power(t, g, krang); got != 7 {
		t.Errorf("Krang power %d, want 7", got)
	}
}

// Another permanent of mine whose ability a draw triggers triggers an
// additional time; an opponent's does not.
func TestKrangTheAllPowerfulDoublesMyDrawTriggersOnly(t *testing.T) {
	g, me, _ := krangTable(t)
	mine := pushCatalogPermanent(g, me.ID, "Chasm Skulker", "Creature — Squid Horror", krangChasmSkulkerOracle, false)
	opp := g.Seats[2]
	theirs := pushCatalogPermanent(g, opp.ID, "Chasm Skulker", "Creature — Squid Horror", krangChasmSkulkerOracle, false)
	krangDraw(t, g, me)
	if got := krangCounters(g, mine); got != 2 {
		t.Errorf("my Skulker: %d counters after one draw, want 2", got)
	}
	krangDraw(t, g, opp)
	if got := krangCounters(g, theirs); got != 1 {
		t.Errorf("their Skulker: %d counters after one draw, want 1 (Krang is not theirs)", got)
	}
}

// A trigger whose cause is not a draw is not doubled.
func TestKrangTheAllPowerfulDoesNotDoubleOtherCauses(t *testing.T) {
	g, me, _ := krangTable(t)
	pridemate := pushCatalogPermanent(g, me.ID, "Ajani's Pridemate", "Creature — Cat Soldier", "95e94dea-5ac0-4d6f-adec-ca147aee861f", false)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 2) })
	mscM12Settle(t, g)
	if got := krangCounters(g, pridemate); got != 1 {
		t.Errorf("Pridemate: %d counters after gaining life, want 1", got)
	}
}

// Without Krang on the battlefield nothing is doubled.
func TestKrangTheAllPowerfulDoublesNothingOnceGone(t *testing.T) {
	g, me, krang := krangTable(t)
	mine := pushCatalogPermanent(g, me.ID, "Chasm Skulker", "Creature — Squid Horror", krangChasmSkulkerOracle, false)
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(krang) })
	krangDraw(t, g, me)
	if got := krangCounters(g, mine); got != 1 {
		t.Errorf("Skulker: %d counters after one draw without Krang, want 1", got)
	}
}
