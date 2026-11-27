package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Memory Sluice — Sorcery {U/B} (Shadowmoor, S82 older sets):
//
//	"Target player mills four cards.
//	 Conspire (As you cast this spell, you may tap two untapped creatures you
//	 control that share a color with it. When you do, copy it and you may choose a new target for the copy.)"
//
// Rotcrown Ghoul's mill primitive as a spell; the copy may name another player.
// Conspire is the engine's optional cost and cast trigger (conspire.go, CR
// 702.78), so the card declares the cost only. The conspire copy is a copy of the spell and may choose a new target (CR 707.10c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "a30b659e-ae2b-4736-9b4f-a9805905bf75",
		Name:          "Memory Sluice",
		Completeness:  effects.CompletenessFull,
		Targets:       effects.TargetPlayer("target player"),
		OptionalCosts: []game.AdditionalCost{effects.Conspire()},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.TargetPlayerMillsEffect(4)(ctx.Game, item)
		},
	})
}
