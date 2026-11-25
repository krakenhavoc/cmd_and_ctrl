package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Revolutionary Rebuff — Instant {1}{U} (Kaladesh, #5539):
//
//	"Counter target nonartifact spell unless its controller pays {2}."
//
// Spell Pierce's counter-unless-paid over a spell that is not an artifact
// spell: the spell's controller is asked to pay as the Rebuff resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "879c8663-5058-4998-9a59-f2e9ea8914dc",
		Name:         "Revolutionary Rebuff",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetSpell("target nonartifact spell", effects.Not(effects.Artifact())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.CounterTheTargetSpellUnlessPaid("Revolutionary Rebuff", "{2}")(ctx.Game, item)
		},
	})
}
