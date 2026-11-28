package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bog Down — Sorcery {2}{B} (Planeshift, S82 OLD b641, #5615):
//
//	"Kicker—Sacrifice two lands. (You may sacrifice two lands in addition to
//	 any other costs as you cast this spell.)
//	 Target player discards two cards. If this spell was kicked, that player
//	 discards three cards instead."
//
// "Instead" is one read of the kick at resolution: the player discards two or
// three, never five. The player picks the cards (CR 701.9a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "8574f57a-768d-42a0-ac8c-cc0a1de181e0",
		Name:          "Bog Down",
		Completeness:  effects.CompletenessFull,
		Targets:       effects.TargetPlayer("target player"),
		OptionalCosts: []game.AdditionalCost{effects.OldB641KickerSacrificeN(2, "two lands", effects.Land())},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			n := 2
			if ctx.WasKicked() {
				n = 3
			}
			return effects.TargetPlayerDiscards(n)(ctx.Game, item)
		},
	})
}
