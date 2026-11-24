package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mind Drain — Sorcery {2}{B} (Zendikar Rising, #5508):
//
//	"Target opponent discards two cards, mills a card, and loses 1 life.
//	 You gain 1 life."
//
// In printed order: the opponent chooses and discards two (CR 701.9b; the
// resolution waits for that prompt), then mills the top card of their
// library, loses 1 life, and the caster gains 1 life. The mill, the loss
// and the gain are skipped if the target player is no longer legal (the
// whole spell fizzles, CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0a88d1b1-ac3c-41da-b3f5-b689af0e45d7",
		Name:         "Mind Drain",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		Purpose:      effects.ForTargets(game.TargetPurpose{Slot: 0, Discards: 2, LifeLoss: 1}),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetPlayer {
						return effects.TargetPlayerDiscards(2)(ctx.Game, ctx.Item)
					}
				}
				return nil
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetPlayer {
						continue
					}
					if err := (effects.MillCards{Player: t.ID, N: 1}).Apply(ctx); err != nil {
						return err
					}
					if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), t.ID, -1); err != nil {
						return err
					}
					return effects.GainLife{Player: ctx.Controller(), Amount: 1}.Apply(ctx)
				}
				return nil
			}),
		),
	})
}
