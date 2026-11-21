package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Malcolm, the Eyes — Legendary Creature — Siren Pirate {U}{R}, 2/2
// (Outlaws of Thunder Junction, #4468):
//
//	"Flying, haste
//	 Whenever you cast your second spell each turn, investigate.
//	 (Create a Clue token. It's an artifact with "{2}, Sacrifice this
//	 token: Draw a card.")"
//
// The trigger is Kraum's YouCastYourSecondSpellEachTurn, and
// investigating is Thraben Inspector's Clue. Only the
// second spell triggers it, so a third that turn makes no Clue.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7932ab35-39dc-4500-aa44-2ad7228f0d2a",
		Name:            "Malcolm, the Eyes",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.YouCastYourSecondSpellEachTurn,
				"Malcolm, the Eyes — investigate",
				effects.Do(effects.Investigate{N: 1})),
		},
	})
}
