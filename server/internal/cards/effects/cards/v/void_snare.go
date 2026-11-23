package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Void Snare — Sorcery {U} (Magic 2015, #5556):
//
//	"Return target nonland permanent to its owner's hand."
//
// Disperse at sorcery speed (bounceTheTargetToHand); a target that left
// in response is skipped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "06e2ce8e-7341-4448-8625-cb1574eb179d",
		Name:         "Void Snare",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target nonland permanent", effects.Nonland()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.BounceTheTargetToHand(ctx.Game, item)
		},
	})
}
