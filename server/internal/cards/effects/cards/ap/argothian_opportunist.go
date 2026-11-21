package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Argothian Opportunist — Creature — Human Scout {2}{G}, 3/2 (The
// Brothers' War, #5482):
//
//	"When this creature enters, create a tapped Powerstone token."
//
// A tapped Powerstone made on enter, as Stone Retrieval Unit's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cc9d3868-9814-4506-80b3-660bc462ff98",
		Name:         "Argothian Opportunist",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Argothian Opportunist — create a tapped Powerstone token",
				effects.Do(effects.CreateTokenAdvanced{Spec: effects.Token(effects.PowerstoneToken()).EntersTapped(), N: 1})),
		},
	})
}
