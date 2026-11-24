package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Golden Ratio — Sorcery {1}{G}{U} (Strixhaven: School of Mages, #5502):
//
//	"Draw a card for each different power among creatures you control."
//
// The number of distinct power values (CR 208.1, read through the layered
// view as the spell resolves, CR 608.2h) among the creatures the caster
// controls. Two 2/2s are one power; a 0-power creature is a power of its
// own. Nothing to count draws nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "724700ad-1e4e-4cbc-af7f-0b3b87ed4bc8",
		Name:         "Golden Ratio",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			seen := map[int]bool{}
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.Controller == ctx.Controller() && c.IsCreature() {
					seen[c.CurrentPower()] = true
				}
			}
			return effects.DrawCards{N: len(seen)}.Apply(ctx)
		},
	})
}
