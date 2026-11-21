package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nimble Thopterist — Creature — Vedalken Artificer {3}{U}, 3/2
// (Aetherdrift, #4464):
//
//	"When this creature enters, create a 1/1 colorless Thopter artifact
//	 creature token with flying."
//
// Broadcast Rambler's enters trigger on a creature: the predefined
// Thopter token, made whether or not the Thopterist is still on the
// battlefield when the trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "01813091-75e8-431b-ba3f-b1ea1ba20a3b",
		Name:         "Nimble Thopterist",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Nimble Thopterist — create a 1/1 colorless Thopter artifact creature token with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Thopter artifact with flying"), N: 1})),
		},
	})
}
