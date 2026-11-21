package testsp

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const personifyOracle = "e1708ad9-35d4-4243-92ca-ddf9d129aa32"

func TestPersonifyIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(personifyOracle)
	if !ok || spec.Name != "Personify" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

// Personify blinks a creature you control back to its owner, without its
// counters, then makes a 1/1 changeling Shapeshifter for you. An
// opponent's creature is not a legal target.
func TestPersonifyFlickersToTheOwnerAndMakesAShapeshifter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.ToMainForCost(t, g)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	borrowed := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Borrowed Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: me.ID})
	effectstest.AddCountersForTest(g, borrowed, game.CounterPlusOne, 2)

	spell := effectstest.RfdHandCard(me, "Personify", "Instant", "{1}{W}", personifyOracle, 0, 0)
	effectstest.FillPoolColored(me, "W", 2)
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(theirs)}); err == nil {
		t.Fatal("Personify targeted a creature an opponent controls")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(borrowed)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	var back game.Card
	found := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Borrowed Bear" {
				back, found = c, true
			}
		}
	})
	if !found {
		t.Fatal("the creature did not return")
	}
	if back.Controller != opp.ID {
		t.Error("the creature returned under its controller's control, not its owner's")
	}
	if back.Counters[game.CounterPlusOne] != 0 {
		t.Error("the returned creature kept its counters")
	}
	var shapes []game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if effects.IsToken(c) && c.Name == "Shapeshifter" {
				shapes = append(shapes, c)
			}
		}
	})
	if len(shapes) != 1 || shapes[0].Controller != me.ID {
		t.Fatalf("Shapeshifter tokens %+v, want one under the caster's control", shapes)
	}
	if !effectstest.HasEffectiveKeyword(t, g, shapes[0].InstanceID, "changeling") {
		t.Error("the Shapeshifter has no changeling")
	}
}

// A token exiled this way doesn't come back (CR 111.8); the Shapeshifter
// is still created.
func TestPersonifyOnATokenLosesItButStillMakesAShapeshifter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMainForCost(t, g)
	tok := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Soldier", TypeLine: "Token Creature — Soldier",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	effectstest.CastCatalogSpell(t, g, "Personify", "Instant", personifyOracle, effectstest.CardTarget(tok))
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Soldier"); n != 0 {
		t.Error("the exiled token came back")
	}
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Shapeshifter"); n != 1 {
		t.Errorf("%d Shapeshifters, want 1", n)
	}
}
