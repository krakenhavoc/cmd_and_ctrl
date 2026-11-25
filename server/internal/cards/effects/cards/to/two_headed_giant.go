package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Two-Headed Giant — Creature — Giant Warrior {2}{R}{R}, 4/4 (Dominaria, S82
// OLD, #5530):
//
//	"Whenever this creature attacks, flip two coins. If both coins come up
//	 heads, this creature gains double strike until end of turn. If both coins
//	 come up tails, this creature gains menace until end of turn."
//
// Two uncalled coins (CR 705.2: only the faces matter), flipped as the trigger
// resolves; one head and one tail does nothing. The keyword is granted to the
// Giant until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e0efb48b-e7e5-4ed0-a7de-c8f72d357319",
		Name:         "Two-Headed Giant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Two-Headed Giant — flip two coins: both heads, double strike; both tails, menace", effects.S82OldB213TwoHeadedGiantFlips),
		},
	})
}
