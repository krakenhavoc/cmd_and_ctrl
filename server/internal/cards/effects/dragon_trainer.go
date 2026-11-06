package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dragon Trainer — Creature — Human {3}{R}{R}, 1/1 (Foundations, #4465):
//
//	"When this creature enters, create a 4/4 red Dragon creature token
//	 with flying."
//
// Prideful Parent's enters trigger with the token table's 4/4 red
// Dragon.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "328a0691-9453-4f2d-a9c3-5c69da252d20",
		Name:         "Dragon Trainer",
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Dragon Trainer — create a 4/4 red Dragon with flying",
				Do(CreateToken{Template: TokenCard("4/4 red Dragon with flying"), N: 1})),
		},
	})
}
