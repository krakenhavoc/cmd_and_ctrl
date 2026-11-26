package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Clinging Mists — Instant {2}{G}:
//
//	"Prevent all combat damage that would be dealt this turn.
//	 Fateful hour — If you have 5 or less life, tap all attacking creatures.
//	 Those creatures don't untap during their controller's next untap step."
//
// The prevention is Fog's. The fateful hour half taps the attackers on the
// table as the spell resolves and freezes exactly those (CR 611.2c).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9be48462-2ec0-4d28-9977-f0232e8a1333",
		Name:         "Clinging Mists",
		Purpose:      game.Purpose{Answers: game.AnswerPrevent},
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := (effects.PreventAllCombatDamageThisTurn{Label: "Clinging Mists: prevent combat damage"}).Apply(ctx); err != nil {
				return err
			}
			if !effects.OldB441FatefulHour(ctx.Game, ctx.Controller()) {
				return nil
			}
			return effects.TapAndFreeze{Targets: effects.AttackingCreatures(ctx.Game), Label: "Clinging Mists — fateful hour"}.Apply(ctx)
		},
	})
}
