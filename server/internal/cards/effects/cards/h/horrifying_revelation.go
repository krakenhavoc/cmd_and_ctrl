package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Horrifying Revelation — Sorcery {B} (Mirrodin Besieged, S82 OLD b465, #5576):
//
//	"Target player discards a card, then mills a card."
//
// In printed order: the player chooses and discards (the resolution waits for
// that prompt, CR 701.9b), then mills the top card of their library (CR
// 701.17a). Mind Drain's shape. A target no longer legal fizzles the whole
// spell (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "36ee4179-83a1-472a-a48c-394c4c0457bc",
		Name:         "Horrifying Revelation",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		Purpose:      effects.ForTargets(game.TargetPurpose{Slot: 0, Discards: 1}),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetPlayer {
						return effects.TargetPlayerDiscards(1)(ctx.Game, ctx.Item)
					}
				}
				return nil
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetPlayer {
						return effects.MillCards{Player: t.ID, N: 1}.Apply(ctx)
					}
				}
				return nil
			}),
		),
	})
}
