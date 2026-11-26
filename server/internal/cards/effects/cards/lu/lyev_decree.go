package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lyev Decree — Sorcery {1}{W} (Dragon's Maze, S82 OLD batch 403, #5564):
//
//	"Detain up to two target creatures your opponents control. (Until your next
//	 turn, those creatures can't attack or block and their activated abilities
//	 can't be activated.)"
//
// Abandon the Post's "up to two" target count with Azorius Arrester's detain
// (CR 701.35a): every target still legal as the spell resolves is detained until
// your next turn, the set pinned now; with no targets chosen the spell does
// nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ace64875-ca4b-45c7-8b46-29ef2f2629dd",
		Name:         "Lyev Decree",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerCombatGrant},
		Targets:      effects.TargetCreature("up to two target creatures your opponents control", effects.OpponentControls()).WithCount(0, 2),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DetainEachLegalTarget("Lyev Decree — detain")(ctx.Game, item)
		},
	})
}
