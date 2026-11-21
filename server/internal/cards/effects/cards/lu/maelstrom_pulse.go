package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Maelstrom Pulse — Sorcery {1}{B}{G} (EDHREC rank 3624):
//
//	"Destroy target nonland permanent and all other permanents with
//	 the same name as that permanent."
//
// The Golgari answer to a token swarm. One target; at resolution
// every other permanent on the battlefield — anyone's, a land
// included ("all other permanents"), a token included — sharing the
// target's name is destroyed with it, each through the single-target
// verb, so an indestructible copy survives while the rest go. The
// name is read at resolution off the effective characteristics, so
// a Clone that copied the target dies with it and a target that
// left in response fizzles the spell.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "95ce305f-34bc-4d6d-b7ba-ffd4b2a25336",
		Name:         "Maelstrom Pulse",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target nonland permanent", effects.Nonland()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.B34DestroyChosenAndAllOthersWithItsName(ctx)
		},
	})
}
