package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const tangleTumblerOracle = "0d81ea20-5e7d-4eea-b0ba-fe688f27df94"

func TestTangleTumblerIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(tangleTumblerOracle)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Tangle Tumbler: %+v %v", spec, ok)
	}
}

func ttBoard(g *game.Game, me *game.Player) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Tangle Tumbler", TypeLine: "Artifact — Vehicle", OracleID: tangleTumblerOracle,
		Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID,
	})
}

func TestTangleTumblerTapsTwoTokensToBecomeAnArtifactCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tumbler := ttBoard(g, me)
	a := fcgToken(g, me.ID, "Token A")
	b := fcgToken(g, me.ID, "Token B")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	c, _ := battlefieldCard(g, tumbler)
	if c.IsCreature() {
		t.Fatal("the Vehicle is a creature before it is animated")
	}
	// A nontoken creature can't pay.
	if err := g.ActivateCatalogAbility(me.ID, tumbler, 1, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{a, bear}}); err == nil {
		t.Error("tapped a nontoken creature to pay")
	}
	if err := g.ActivateCatalogAbility(me.ID, tumbler, 1, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{a, b}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	c, _ = battlefieldCard(g, tumbler)
	if !c.IsCreature() || !c.IsArtifact() {
		t.Errorf("types %v, want an artifact creature", c.Effective().Types)
	}
	ca, _ := battlefieldCard(g, a)
	cb, _ := battlefieldCard(g, b)
	if !ca.Tapped || !cb.Tapped {
		t.Error("the tokens were not tapped")
	}
}

func TestTangleTumblerNeedsTwoTokens(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tumbler := ttBoard(g, me)
	a := fcgToken(g, me.ID, "Token A")
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, tumbler, 1, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{a}}); err == nil {
		t.Error("animated with one token")
	}
}

func TestTangleTumblerPutsACounterOnATargetCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tumbler := ttBoard(g, me)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	floatForTest(g, me, "GGG")
	if err := g.ActivateCatalogAbility(me.ID, tumbler, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters: %d, want 1", got)
	}
}
