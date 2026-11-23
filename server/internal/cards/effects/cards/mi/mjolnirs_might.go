package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mjölnir's Might — Sorcery {3}{R} (Marvel Super Heroes Commander,
// #4455):
//
//	"Mjölnir's Might deals 4 damage to target player.
//	 Exile the top card of your library. Until the end of your next turn,
//	 you may play that card."
//
// Two instructions in printed order. The damage goes to the chosen
// player; the impulse exile is the caster's own top card with the
// shared window of "until the end of your next turn"
// (ExileTopNUntilYourNextTurn). If the target is illegal on resolution
// the whole spell does nothing (CR 608.2b), exile included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8c944733-3fb5-4d1f-af1a-3fd31377eb9a",
		Name:         "Mjölnir's Might",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 4)),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetPlayer {
						return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 4}.Apply(ctx)
					}
				}
				return nil
			}),
			effects.Instruction(func(ctx *effects.Context) error { return effects.ExileTopNUntilYourNextTurn(ctx, 1) }),
		),
	})
}
