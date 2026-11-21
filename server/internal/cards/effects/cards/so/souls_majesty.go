package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soul's Majesty — Sorcery {4}{G} (EDHREC rank 3312):
//
//	"Draw cards equal to the power of target creature you control."
//
// Green's card draw off a big body. The power is read at resolution
// — current power, so counters and anthems count — from the creature
// chosen at announce; a creature that left in response counters the
// spell (CR 608.2b) and nothing is drawn, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "50fd4cf3-c347-4eb9-ab15-a9b0c5ea8b0f",
		Name:         "Soul's Majesty",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				return effects.DrawCards{Player: item.Controller, N: effects.B31CurrentPowerOf(ctx.Game, t.ID)}.Apply(ctx)
			}
			return nil
		},
	})
}
