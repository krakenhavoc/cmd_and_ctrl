package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Djeru's Renunciation — Instant {1}{W} (Hour of Devastation, #5534):
//
//	"Tap up to two target creatures.
//	 Cycling {W} ({W}, Discard this card: Draw a card.)"
//
// Repel the Darkness's tap without its draw (the shared tapEachLegalTarget,
// which skips a target that left), and the shared Cycling row.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "11b21086-0779-4d16-8a56-252c6bee9bed",
		Name:         "Djeru's Renunciation",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve:    effects.ResolveSteps(effects.Instruction(func(ctx *effects.Context) error { return effects.TapEachLegalTarget(ctx.Game, ctx.Item) })),
		Activated:    []effects.ActivatedAbility{effects.Cycling("{W}")},
	})
}
