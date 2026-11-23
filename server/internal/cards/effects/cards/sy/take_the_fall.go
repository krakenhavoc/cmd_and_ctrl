package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Take the Fall — Instant {U} (Outlaws of Thunder Junction, #4468):
//
//	"Target creature gets -1/-0 until end of turn. It gets -4/-0 until end
//	 of turn instead if you control an outlaw. (Assassins, Mercenaries,
//	 Pirates, Rogues, and Warlocks are outlaws.)
//	 Draw a card."
//
// Depower's body with the size read as the spell resolves: Outlaws' Fury's
// outlaw check (any permanent you control with one of the five outlaw
// types) picks -4/-0 over -1/-0, one or the other, never both. The draw is
// its own sentence and happens after the shrink.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "279f3aa5-5807-4797-8cd1-e009984a3691",
		Name:         "Take the Fall",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, -1, 0)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if t, ok := effects.B16FirstLegalTargetCard(ctx); ok {
				power, label := -1, "Take the Fall — -1/-0"
				if effects.B40ControlsA(ctx.Game, ctx.Controller(), effects.Outlaw()) {
					power, label = -4, "Take the Fall — -4/-0 (you control an outlaw)"
				}
				if err := (effects.BoostUntilEOT{Target: t, Power: power, Label: label}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
		},
	})
}
