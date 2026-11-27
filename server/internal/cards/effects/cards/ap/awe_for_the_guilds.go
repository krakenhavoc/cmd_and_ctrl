package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Awe for the Guilds — Sorcery {2}{R} (Dragon's Maze, S82 OLD batch 403, #5564):
//
//	"Monocolored creatures can't block this turn."
//
// Goblin Locksmith's query-filtered "can't block this turn" (CR 509.1b) over every
// creature of exactly one colour, every player's, read live: a creature that
// arrives later in the turn, or that gains or loses a colour, is judged as
// blockers are declared (CR 611.2c: it changes no characteristic). The one-colour
// test reads effective colours (CR 105.2b), so a colourless or a multicoloured
// creature may still block.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cb50ec2e-cb27-4ee5-a7de-64ab93619cea",
		Name:         "Awe for the Guilds",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerCombatGrant},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CantBlockThisTurnMatching(false, "Awe for the Guilds — monocolored creatures can't block",
				game.PermanentQuery{Types: []string{"creature"}, Monocolored: true}).Apply(ctx)
		},
	})
}
