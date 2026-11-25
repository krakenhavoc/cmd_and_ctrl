package testss82m

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	oldB162ScaleUpOracle = "2fd025ef-14f5-4621-831d-dd12b7c604dd"
)

func oldB162Bear(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bear", Colors: []string{"W"},
		Power: 2, Toughness: 2, Keywords: []string{"vigilance"}, Owner: owner, Controller: owner,
	})
}

func TestOldB162ScaleUpMakesOneCreatureAGreenWurm(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := oldB162Bear(g, me.ID, "Bear A")
	b := oldB162Bear(g, me.ID, "Bear B")
	effectstest.CastCatalogSpell(t, g, "Scale Up", "Sorcery", oldB162ScaleUpOracle, []game.TargetRef{{Kind: game.TargetCard, ID: a}})
	effectstest.PassPriorityAroundTable(t, g)
	ch := effectstest.EffectiveCardForTest(t, g, a).Effective()
	if !reflect.DeepEqual(ch.Subtypes, []string{"Wurm"}) || !reflect.DeepEqual(ch.Colors, []string{"G"}) {
		t.Errorf("the target is %v %v, want a green Wurm", ch.Colors, ch.Subtypes)
	}
	if p, tough := effectstest.EffectivePower(t, g, a), effectstest.EffectiveToughness(t, g, a); p != 6 || tough != 4 {
		t.Errorf("the target is %d/%d, want 6/4", p, tough)
	}
	if p := effectstest.EffectivePower(t, g, b); p != 2 {
		t.Errorf("the other creature has power %d, want 2", p)
	}
	effectstest.AdvanceTo(t, g, game.StepCleanup)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, a); p != 2 {
		t.Errorf("the effect outlived the turn: power %d", p)
	}
}

func TestOldB162ScaleUpOverloadedChangesEveryCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	a := oldB162Bear(g, me.ID, "Bear A")
	b := oldB162Bear(g, me.ID, "Bear B")
	theirs := oldB162Bear(g, opp.ID, "Their Bear")
	effectstest.CastWithAltCost(t, g, "Scale Up", "Sorcery", oldB162ScaleUpOracle, "overload")
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if p, tough := effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id); p != 6 || tough != 4 {
			t.Errorf("your creature is %d/%d, want 6/4", p, tough)
		}
	}
	if p := effectstest.EffectivePower(t, g, theirs); p != 2 {
		t.Errorf("their creature has power %d, want 2", p)
	}
}
