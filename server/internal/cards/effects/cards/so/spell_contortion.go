package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spell Contortion — Instant {2}{U} (Worldwake, S82 OLD b488):
//
//	"Multikicker {1}{U} (You may pay an additional {1}{U} any number of times
//	 as you cast this spell.)
//	 Counter target spell unless its controller pays {2}. Draw a card for each
//	 time Spell Contortion was kicked."
//
// Crush Dissent's counter-unless-paid prompt, with the draw as the second step: it waits
// for the payer's answer (CR 608.2c) and is made whether or not the spell was countered,
// once for each time the card was kicked (CR 702.33d), none when it was not. The
// controller of the targeted spell, not you, decides whether to pay.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "01003192-8927-4e55-8dd9-e9e088a3fbe4",
		Name:          "Spell Contortion",
		Completeness:  effects.CompletenessFull,
		Purpose:       game.Purpose{Draws: 1},
		OptionalCosts: []game.AdditionalCost{effects.Multikicker("{1}{U}", 20)},
		Targets:       effects.TargetSpell("target spell"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.CounterTheTargetSpellUnlessPaid("Spell Contortion", "{2}")(ctx.Game, ctx.Item)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.DrawCards{N: ctx.KickedTimes()}.Apply(ctx)
			}),
		),
	})
}
