package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Forfend — Instant {1}{W}:
//
//	"Prevent all damage that would be dealt to creatures this turn."
//
// #2045: the not-one-use shield with no source, protecting every
// creature, whoever controls it, and no player. The set is read as the
// damage would be dealt (CR 611.2c): a creature that enters after
// Forfend resolved is protected, and a permanent that stops being a
// creature is not.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "59f8f2c4-d30d-411c-9fac-a5a3219bb124",
		Name:         "Forfend",
		Purpose:      game.Purpose{Answers: game.AnswerPrevent}, // #4734 slice 6
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.SourceShieldSpell(effects.PreventDamageFromSource{Protect: effects.ShieldCreatures}),
	})
}
