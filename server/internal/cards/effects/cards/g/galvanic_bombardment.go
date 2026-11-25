package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Galvanic Bombardment — Instant {R} (Eldritch Moon, S82 OLD, #5541):
//
//	"Galvanic Bombardment deals X damage to target creature, where X is 2
//	 plus the number of cards named Galvanic Bombardment in your
//	 graveyard."
//
// X is counted as the spell resolves (CR 608.2h); the spell itself is on the
// stack then, not in the graveyard, so it does not count itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "13d8347f-adb9-41b5-a7b0-a6153bcf2d6c",
		Name:         "Galvanic Bombardment",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 2)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			x := 2 + effects.S82OldB276NamedInGraveyard(ctx.Game, ctx.Controller(), "Galvanic Bombardment")
			return effects.SourceDealsDamageToEachLegalTarget(x)(ctx.Game, item)
		},
	})
}
