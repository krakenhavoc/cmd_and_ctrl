package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flick a Coin — Instant {2}{R}:
//
//	"Flick a Coin deals 1 damage to any target. You create a Treasure
//	 token.
//	 Draw a card."
//
// Three unconditional sentences in printed order. The Treasure and the
// card are not tied to the damage: a target that merely took the
// damage and died changes nothing here, while a spell whose only
// target became illegal is countered on resolution (CR 608.2b) and
// does none of it. (#3213, the Kuja deck.)
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "990583ef-5bf4-4e14-a8f1-61d32c6a75df",
		Name:         "Flick a Coin",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 1)),
		// The Treasure and the card wait for a prompt the damage asks (CR 608.2c, #4155).
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.DamageToFirstTargetStep(1)(ctx.Item, ctx) }),
			effects.CreateToken{Template: effects.TreasureToken(), N: 1},
			effects.DrawCards{N: 1},
		),
	})
}
