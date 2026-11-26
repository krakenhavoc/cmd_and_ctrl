package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flesh // Blood — split card, oracle 0741151a-fec0-4ed3-9295-0d0f56d24165
// (Dragon's Maze, S82 OLD batch 402, #5564):
//
//	Flesh — Sorcery {3}{B}{G}: "Exile target creature card from a graveyard.
//	        Put X +1/+1 counters on target creature, where X is the power of the
//	        card you exiled."
//	Blood — Sorcery {R}{G}: "Target creature you control deals damage equal to
//	        its power to any target."
//	        Fuse (You may cast one or both halves of this card from your hand.)
//
// Both halves are ADR 0034 faces (Flesh under the bare oracle ID, Blood under
// "#1"); the fused cast is the engine's (CR 702.102), each half reading its own
// targets.
//
// Flesh reads the card's power as it is exiled (the power it had in the
// graveyard) and puts that many counters on the second target if it is still a
// legal creature; with the card gone from the graveyard nothing was exiled and X
// is zero. Blood is Bite Down's one-sided fight with a free second slot
// (S82OldB402OneSidedBiteAtAnyTarget): slot 0 deals, its power read as it
// resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0741151a-fec0-4ed3-9295-0d0f56d24165",
		Name:         "Flesh",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCardInGraveyard("target creature card from a graveyard", effects.Creature()),
			effects.TargetCreature("target creature"),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			card, ok := ctx.ClauseTarget(0)
			if !ok || card.Kind != game.TargetCard {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(card.ID)
			if !ok {
				return nil
			}
			return effects.ExileTarget{
				Target:   card.ID,
				ThenTail: effects.WithN(ctx.ItemTail(fleshCounters), c.CurrentPower()),
			}.Apply(ctx)
		},
	})
	effects.Register(effects.Spec{
		OracleID:     "0741151a-fec0-4ed3-9295-0d0f56d24165#1",
		Name:         "Blood",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetAny(),
		),
		OnResolve: effects.S82OldB402OneSidedBiteAtAnyTarget,
	})
}

// fleshCounters is "put X +1/+1 counters on target creature, where X is the power
// of the card you exiled", run once the exile has settled: X is the tail's N (the
// card's power, read before it moved), the tail is told the card only if it
// reached exile, and the counters go on the second target if it is still a legal
// creature (CR 608.2b). An on-disk key: never renamed, never reused.
var fleshCounters = effects.CardTail("flesh-blood/flesh-counters", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 || o.Tail.N <= 0 {
		return nil
	}
	creature, ok := ctx.ClauseTarget(1)
	if !ok || creature.Kind != game.TargetCard {
		return nil
	}
	return effects.AddCounter{Target: creature.ID, Kind: game.CounterPlusOne, N: o.Tail.N}.Apply(ctx)
})
