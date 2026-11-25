package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bonus Round — Sorcery {1}{R}{R} (Battlebond):
//
//	"Until end of turn, whenever a player casts an instant or sorcery spell,
//	 that player copies it and may choose new targets for the copy."
//
// A repeating delayed trigger (CR 603.7b) on every player's casts, Doublecast's
// copy with the caster, not Bonus Round's controller, as the copy's controller
// and the one who chooses new targets (CR 707.10). A copy is created, not cast
// (CR 707.10), so it does not trigger Bonus Round again. The trigger ends at
// cleanup (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "40ab2b22-9cf6-4a73-a05f-6ce496d10bf7",
		Name:         "Bonus Round",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DelayedOnEvent{
				Label:      "Bonus Round — that player copies it and may choose new targets",
				On:         []game.EventKind{game.EventCast},
				Condition:  effects.OldB206AnyPlayerCastsCondition,
				CondParams: game.EffectParams{Filter: game.CastFilter{Types: []string{"Instant", "Sorcery"}}},
				Body:       effects.OldB206CopyTheSpellForItsCasterBody,
				Repeats:    true,
			}.Apply(ctx)
		},
	})
}
