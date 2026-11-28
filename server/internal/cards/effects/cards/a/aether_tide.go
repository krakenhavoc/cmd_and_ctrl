package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aether Tide — Sorcery {X}{U} (Exodus, S82 OLD batch 742, #5626):
//
//	"As an additional cost to cast this spell, discard X creature cards.
//	 Return X target creatures to their owners' hands."
//
// X is announced with the cast (CR 107.3a) and is both the {X} in the cost
// and the number of creature cards discarded as it is paid (DiscardXCreatureCost,
// Channeled Force's discard); exactly X targets (CountFromX), each re-checked
// as the spell resolves (CR 608.2b) and returned if still legal.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "2fbf95b4-bcf4-4b5e-b5dc-0294f2b48d3e",
		Name:           "Aether Tide",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.DiscardXCreatureCost(),
		Targets:        effects.TargetsCountedByX(effects.TargetCreature("X target creatures")),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.NeoB20BounceTargetedPermanents(ctx)
		},
	})
}
