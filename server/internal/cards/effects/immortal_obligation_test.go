package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const immortalObligationOracle = "b34080fe-69d1-4cc9-9feb-c37762f26a23"

func TestImmortalObligationIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(immortalObligationOracle)
	if !ok || spec.Name != "Immortal Obligation" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("registration: %+v", spec)
	}
}

func TestImmortalObligationReturnsUnderTheOpponentsControlBoundToTheDutyCounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	var walker uuid.UUID
	g.WithWriteLock(func() {
		walker = pushPermanent(g, me.ID, game.Card{Name: "My Walker", TypeLine: "Legendary Planeswalker — Test",
			Counters: map[string]int{game.CounterLoyalty: 3}})
	})
	dead := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	mine := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: mine, Name: "My Dead Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	myBear := b12Creature(g, me.ID, "My Attacker", "Creature — Bear", 2, 2)
	blocker := b12Creature(g, opp.ID, "Other Blocker", "Creature — Wall", 0, 4)

	// Only a creature card in an OPPONENT's graveyard is a legal target.
	spell := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: spell, Name: "Immortal Obligation", TypeLine: "Instant", OracleID: immortalObligationOracle, ManaCost: "{1}{W}", Owner: me.ID, Controller: me.ID})
	me.ManaPool.AddMana(game.ManaToken{Color: "W"}, game.ManaToken{Color: "C"})
	err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine}}})
	if err == nil {
		t.Fatal("targeted a card in my own graveyard")
	}

	tdCastFromHand(t, g, game.Card{Name: "Immortal Obligation", TypeLine: "Instant", OracleID: immortalObligationOracle, ManaCost: "{1}{W}"},
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: dead}}})
	passPriorityAroundTable(t, g)

	c, ok := g.LookupCardForEffect(dead)
	if !ok || !g.Battlefield.Contains(dead) {
		t.Fatal("the creature did not return")
	}
	if c.Controller != opp.ID {
		t.Errorf("controller = %v, want the opponent", c.Controller)
	}
	if countersOn(g, dead, "duty") != 1 {
		t.Error("no duty counter")
	}
	if !c.Goaded() || !slices.Contains(c.AllGoaders(), me.ID) {
		t.Error("not goaded")
	}

	// It can't attack me or my planeswalker, but may attack another player.
	advanceToDeclareAttackersOf(t, g, 1)
	for _, target := range []uuid.UUID{me.ID, walker} {
		if err := g.Clone().DeclareAttacker(dead, target); err == nil {
			t.Errorf("it attacked %s", target)
		}
	}
	if err := g.Clone().DeclareAttacker(dead, third.ID); err != nil {
		t.Errorf("attacking another player: %v", err)
	}

	// It can't block my creatures, but another creature can.
	endTurn(t, g)
	advanceToDeclareAttackersOf(t, g, 0)
	if err := g.DeclareAttacker(myBear, opp.ID); err != nil {
		t.Fatalf("my attack: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(dead, myBear); err == nil {
		t.Error("the returned creature blocked its goader's creature")
	}
	if err := g.DeclareBlocker(blocker, myBear); err != nil {
		t.Errorf("another creature could not block: %v", err)
	}

	// CR 611.2b: with the counter gone every rule ends.
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(dead, "duty", -1)
		g.RecomputeLayersIfStaleLocked()
	})
	after, _ := g.LookupCardForEffect(dead)
	if after.Goaded() {
		t.Error("still goaded without a duty counter")
	}
}
