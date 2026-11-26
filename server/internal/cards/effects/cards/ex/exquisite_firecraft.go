package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Exquisite Firecraft — Sorcery {1}{R}{R} (Magic Origins, #5549):
//
//	"Exquisite Firecraft deals 4 damage to any target.
//	 Spell mastery — If there are two or more instant and/or sorcery cards in
//	 your graveyard, this spell can't be countered."
//
// Lightning Strike for 4 with Banefire's conditional "can't be countered",
// judged when something tries to counter it: the spell is on the stack
// then, so it does not count itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e83c617f-1e07-40db-9b12-ae01122a3e24",
		Name:         "Exquisite Firecraft",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 4)), Answers: game.AnswerRemove},
		CantBeCounteredIf: func(g *game.Game, item *game.StackItem) bool {
			return effects.SpellMastery(g, item.Controller)
		},
		OnResolve: effects.DamageToFirstTargetStep(4),
	})
}
