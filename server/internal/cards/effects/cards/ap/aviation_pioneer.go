package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aviation Pioneer — Creature — Human Artificer {2}{U}, 1/2
// (Core Set 2019, #5527):
//
//	"When this creature enters, create a 1/1 colorless Thopter artifact creature token with flying."
//
// Nimble Thopterist's text: one Thopter token as it enters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ad11ff81-acaa-4529-ba15-718dadbea259",
		Name:         "Aviation Pioneer",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Aviation Pioneer — create a 1/1 colorless Thopter artifact creature token with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Thopter artifact with flying"), N: 1})),
		},
	})
}
