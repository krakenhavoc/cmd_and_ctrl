package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lethal Sting — Sorcery {2}{B}: "As an additional cost to cast this
// spell, put a -1/-1 counter on a creature you control. Destroy target
// creature." (Hour of Devastation, #5948). One of the cards Pinchy
// McStingbutt conjures (ADR 0099 amendment 2026-11-24).
//
// The cost is the action CR 701.68a calls blight 1 ("put a -1/-1 counter
// on a creature you control"), paid by the same component (blight_ids):
// the caster names the creature, a cost's counter is not doubled, and
// with no creature to put it on the spell can't be cast. The card does
// not say "blight", and nothing in the engine reads the word, so the
// label is the printed clause.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "464749e1-9df2-47d0-8fcd-516f566022b5",
		Name:           "Lethal Sting",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.MinusOneCounterOnACreatureYouControlCost(),
		Purpose:        game.Purpose{Answers: game.AnswerRemove},
		Targets:        effects.TargetCreature("target creature"),
		OnResolve:      effects.ResolveSteps(effects.Instruction(effects.KineticOozeDestroy)),
	})
}
