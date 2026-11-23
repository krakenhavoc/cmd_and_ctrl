package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wingblade Disciple — Creature — Human Monk {2}{U}, 2/2 (Tarkir:
// Dragonstorm, #4463):
//
//	"Flying
//	 Flurry — Whenever you cast your second spell each turn, create a 1/1
//	 white Bird creature token with flying."
//
// Flying is a printed keyword. Flurry is an ability word (CR 207.2c): the
// trigger is YouCastYourSecondSpellEachTurn, Cori-Steel Cutter's
// condition, so any second spell counts, creature or not, and the first
// and third do nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c566317e-cd82-46c1-b506-f41256d815f6",
		Name:            "Wingblade Disciple",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.YouCastYourSecondSpellEachTurn,
				"Wingblade Disciple — flurry: create a 1/1 white Bird creature token with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Bird with flying"), N: 1})),
		},
	})
}
