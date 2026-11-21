package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thwart the Enemy — Instant {2}{G}:
//
//	"Prevent all damage that would be dealt this turn by creatures your
//	 opponents control."
//
// #2026's controller test, relative to the spell's controller, read as
// each creature would deal damage (CR 609.7b): combat damage or not, to
// anything. A creature an opponent gains control of later this turn is
// caught, and one you take from them is not.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5799baed-3457-4bf2-adf3-239a41dcc1c8",
		Name:         "Thwart the Enemy",
		Purpose:      game.Purpose{Answers: game.AnswerPrevent}, // #4734 slice 6
		Completeness: effects.CompletenessFull,
		OnResolve: effects.SourceShieldSpell(effects.PreventDamageFromSource{Protect: effects.ShieldAnything, Queries: effects.CreatureSources(),
			Filter: game.DamageSourceFilter{Controller: game.SourceControllerOpponents}}),
	})
}
