package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mirror Strike — Instant {3}{W}:
//
//	"All combat damage that would be dealt to you this turn by target
//	 unblocked creature is dealt to its controller instead."
//
// ADR 0108 §9 (#1905): a redirection for the rest of the turn of the
// target's combat damage to you, dealt to its controller as it is when
// the damage would be dealt.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "184fdd23-8235-4500-b2f5-d3b0b93a8f35",
		Name:         "Mirror Strike",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target unblocked creature", effects.UnblockedAttacker()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			return effects.RedirectDamage{From: t.ID, Protect: effects.ShieldYou, CombatOnly: true, To: effects.RedirectToSourceController}.Apply(ctx)
		},
	})
}
