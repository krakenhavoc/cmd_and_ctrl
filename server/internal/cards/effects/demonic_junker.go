package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Demonic Junker — Artifact — Vehicle {6}{B}, 4/3 (Aetherdrift, #4464):
//
//	"Affinity for artifacts
//	 When this Vehicle enters, for each player, destroy up to one target
//	 creature that player controls. If a creature you controlled was
//	 destroyed this way, put two +1/+1 counters on this Vehicle.
//	 Crew 2"
//
// Affinity is AffinityFor over artifacts. The enters trigger is Molten
// Primordial's per-player clauses (here every player, the controller
// included, each "up to one target creature that player controls"), one
// destroy of the chosen set (CR 701.8), and a keyed tail told the
// creatures that actually went to a graveyard as they were last known
// (CR 603.10): the two counters follow when one of them was YOURS, so an
// indestructible creature of yours, or one a replacement exiled, doesn't
// count. Crew 2 is the ordinary crew row.
//
// No simplification.
func init() {
	enters := WhenThisEnters("Demonic Junker — for each player, destroy up to one target creature that player controls", demonicJunkerDestroy)
	enters.TargetsFrom = demonicJunkerClauses
	Register(Spec{
		OracleID:     "88075828-a76e-412d-b738-182f50e3a133",
		Name:         "Demonic Junker",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			AffinityFor("Affinity for artifacts", Artifact()),
		},
		Triggered: []game.TriggeredAbility{enters},
		Activated: []ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Demonic Junker"),
		}},
	})
}

// demonicJunkerClauses is one "up to one target creature <player>
// controls" clause per player still in the game, in seat order. It reads
// nothing the game can change but the seats, as Molten Primordial's does.
func demonicJunkerClauses(_ game.TriggerContext, _ *game.Card, g *game.Game) *game.TargetSpec {
	var clauses []*game.TargetSpec
	for _, p := range g.Seats {
		if p == nil || p.Eliminated {
			continue
		}
		clauses = append(clauses, TargetCreature("up to one target creature "+p.Name+" controls",
			ControlledBy(p.ID)).WithCount(0, 1))
	}
	if len(clauses) == 0 {
		return nil
	}
	return Clauses(clauses[0], clauses[1:]...)
}

func demonicJunkerDestroy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			ids = append(ids, t.ID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return g.DestroyPermanentsKeyedForEffect(ids, ctx.Tail(demonicJunkerCounters))
}

// demonicJunkerCounters is "if a creature you controlled was destroyed
// this way, put two +1/+1 counters on this Vehicle".
var demonicJunkerCounters = MassTail("mass/demonic-junker-counters", func(ctx *Context, swept []game.Card, _ int) error {
	for _, c := range swept {
		if c.Controller == ctx.Controller() {
			return AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
		}
	}
	return nil
})
