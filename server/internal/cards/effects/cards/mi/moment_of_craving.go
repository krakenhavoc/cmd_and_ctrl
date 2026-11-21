package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Moment of Craving — Instant {1}{B} (Foundations, #4465):
//
//	"Target creature gets -2/-2 until end of turn. You gain 2 life."
//
// Gollum's Bite's shrink with Tamiyo's Safekeeping's life rider: the
// still-legal target gets -2/-2 (boostTargetUntilEOT), then the caster
// gains 2 life. With its only target gone the spell does not resolve
// at all, so the life gain needs no guard of its own.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1489551f-5714-471b-a0e6-f40bc3085afc",
		Name:         "Moment of Craving",
		Purpose:      game.Purpose{LifeGain: 2, Targets: game.ForTargets(effects.HarmUntilEOT(0, -2, -2))},
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.BoostTargetUntilEOT("Moment of Craving — -2/-2", -2, -2)(ctx.Game, item); err != nil {
				return err
			}
			return effects.GainLife{Player: ctx.Controller(), Amount: 2}.Apply(ctx)
		},
	})
}
