package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scorched Earth — {X}{R} Sorcery (Tempest, S82 OLD batch 769, #5628):
//
//	"As an additional cost to cast this spell, discard X land cards.
//	 Destroy X target lands."
//
// Immoral Bargain's X targets with a land discard for the cost
// (DiscardXLandCost): the X announced with the cast is the number of land cards
// discarded and of lands targeted, and a nonland card never pays it. The lands
// are destroyed as one event (VsDestroyTheTargets); a target that has become
// illegal is skipped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "fd1ce3e0-9c2d-48b6-9a14-4e71921cdfa4",
		Name:           "Scorched Earth",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.DiscardXLandCost(),
		Targets:        effects.VsXTarget("X target lands", effects.Land()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.VsDestroyTheTargets(ctx)
		},
	})
}
