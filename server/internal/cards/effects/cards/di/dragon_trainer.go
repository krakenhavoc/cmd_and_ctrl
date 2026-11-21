package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "328a0691-9453-4f2d-a9c3-5c69da252d20",
		Name:         "Dragon Trainer",
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Dragon Trainer — create a 4/4 red Dragon with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("4/4 red Dragon with flying"), N: 1})),
		},
	})
}
