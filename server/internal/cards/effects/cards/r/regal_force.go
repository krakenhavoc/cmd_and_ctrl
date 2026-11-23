package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Regal Force — Creature — Elemental {4}{G}{G}{G}, 5/5 (EDHREC rank
// 3605):
//
//	"When this creature enters, draw a card for each green creature
//	 you control."
//
// The seven-drop that refills the hand. Counted as the trigger
// resolves, off the effective colours, so a creature something else
// painted green counts and the Force itself — green — counts,
// as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e264ffe3-0252-49d6-b990-dbb3654325a5",
		Name:         "Regal Force",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Regal Force — draw a card for each green creature you control", effects.B34DrawPerCreatureYouControlMatching(effects.B34IsGreen)),
		},
	})
}
