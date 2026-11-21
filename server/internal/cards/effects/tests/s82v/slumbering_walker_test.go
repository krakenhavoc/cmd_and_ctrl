package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const slumberingWalkerOracle = "6d972501-4174-4bf0-a26b-2edef31b3a0c"

func TestSlumberingWalkerIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(slumberingWalkerOracle)
	if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

// slumberingWalkerGraveyard seeds a creature card with power 2, one with
// power 3 and a noncreature card into p's graveyard.
func slumberingWalkerGraveyard(p *game.Player) (small, big, rock uuid.UUID) {
	s := game.Card{InstanceID: uuid.New(), Name: "Small Card", TypeLine: "Creature — Elf", Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID}
	b := game.Card{InstanceID: uuid.New(), Name: "Big Card", TypeLine: "Creature — Giant", Power: 3, Toughness: 3, Owner: p.ID, Controller: p.ID}
	r := game.Card{InstanceID: uuid.New(), Name: "Rock Card", TypeLine: "Artifact", Owner: p.ID, Controller: p.ID}
	p.Graveyard.PushTop(s)
	p.Graveyard.PushTop(b)
	p.Graveyard.PushTop(r)
	return s.InstanceID, b.InstanceID, r.InstanceID
}

// It enters with two -1/-1 counters; at your end step removing one
// returns a creature card with power 2 or less, chosen as the reflexive
// trigger goes on the stack.
func TestSlumberingWalkerRemovesACounterToReturnASmallCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	small, big, rock := slumberingWalkerGraveyard(me)
	walker := effectstest.RfdCast(t, g, "Slumbering Walker", "Creature — Giant Warrior", slumberingWalkerOracle, 4, 7)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtB9Counters(g, walker, game.CounterMinusOne); got != 2 {
		t.Fatalf("-1/-1 counters on entry = %d, want 2", got)
	}
	if p := effectstest.EffectivePowerWithCounters(t, g, walker); p != 2 {
		t.Fatalf("power %d with two -1/-1 counters, want 2", p)
	}

	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("a target was asked before the counter came off")
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no reflexive return target prompt")
	}
	for _, bad := range []uuid.UUID{big, rock} {
		if err := g.ResolvePickTargets(p.ID, me.ID, effectstest.CardRefs(bad)); err == nil {
			t.Fatalf("the reflexive trigger targeted %s", bad)
		}
	}
	effectstest.PickTargetRefs(t, g, me.ID, nil, effectstest.CardRefs(small)...)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(small) {
		t.Error("the power 2 creature card did not return to the battlefield")
	}
	if got := effectstest.TmtB9Counters(g, walker, game.CounterMinusOne); got != 1 {
		t.Errorf("-1/-1 counters after the removal = %d, want 1", got)
	}
}

// Declining removes nothing and returns nothing; a Walker with no
// counters returns nothing even when you say yes.
func TestSlumberingWalkerDeclinedOrWithoutCountersReturnsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	small, _, _ := slumberingWalkerGraveyard(me)
	walker := effectstest.RfdCast(t, g, "Slumbering Walker", "Creature — Giant Warrior", slumberingWalkerOracle, 4, 7)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || g.Battlefield.Contains(small) {
		t.Fatal("declining still returned a card")
	}
	if got := effectstest.TmtB9Counters(g, walker, game.CounterMinusOne); got != 2 {
		t.Fatalf("-1/-1 counters after declining = %d, want 2", got)
	}

	g.WithWriteLock(func() { _ = g.AddCounterForEffect(walker, game.CounterMinusOne, -2) })
	effectstest.AdvanceToStepOf(t, g, (seat+1)%len(g.Seats), game.StepUpkeep)
	effectstest.AdvanceToStepOf(t, g, seat, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || g.Battlefield.Contains(small) {
		t.Error("a Walker with no counters returned a card")
	}
}
