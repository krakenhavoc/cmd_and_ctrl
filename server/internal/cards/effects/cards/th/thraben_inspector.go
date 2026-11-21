package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thraben Inspector — 1/2 Creature — Human Soldier for {W}:
//
//	"When Thraben Inspector enters the battlefield, investigate."
//
// Investigating is "create a Clue token", and the Clue is where the
// value is: a card in artifact form that costs {2} to cash in. The
// whole loop — make the token, crack it later for the draw — only
// became possible in S21, because a Clue is nothing but its
// activated ability.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "caa02547-66e3-4e27-a2d3-5e94f3e7a069",
		Name:         "Thraben Inspector",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Thraben Inspector — investigate", effects.Do(effects.Investigate{N: 1})),
		},
	})
}
