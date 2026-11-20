package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Discombobulate — Instant {2}{U}{U}:
//
//	"Counter target spell.
//	 Look at the top four cards of your library, then put them back
//	 in any order."
//
// Counterspell's shape plus LookAtTop, the S22 primitive built for
// exactly "look at N, then reorder" (Ponder, Sensei's Divining Top).
// Nothing follows the reorder on this card — no shuffle branch, no
// draw riding its Then — so unlike Ponder's chained prompts, the two
// sentences are just sequential: counter first (synchronous), then
// queue the look-at-top prompt. LookAtTop.Apply only QUEUES the
// prompt; nothing moves until the caster answers it, which is
// correct here because there's nothing after it in the text to get
// the ordering trap wrong.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b58d7a20-4bcd-4c33-8cda-955362525f48",
		Name:         "Discombobulate",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			if err := (effects.CounterTarget{StackID: item.Targets[0].ID}).Apply(ctx); err != nil {
				return err
			}
			return effects.LookAtTop{Player: ctx.Controller(), N: 4}.Apply(ctx)
		},
	})
}
