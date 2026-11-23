package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Master Splicer — Creature — Phyrexian Human Artificer {3}{W}, 1/1
// (New Phyrexia, reprinted in Commander decks, #5908):
//
//	"When this creature enters, create a 3/3 colorless Phyrexian Golem
//	 artifact creature token.
//	 Golems you control get +1/+1."
//
// The anthem reads effective subtypes through the layers, so it lifts the
// token Golem (and any Golem you control, tokens or cards), not the Splicer.
//
// No simplification.
func init() {
	golems := effects.TribeFilter{Tribes: []string{"Golem"}, YoursOnly: true}
	effects.Register(effects.Spec{
		OracleID:     "9e0c5919-aed6-4d98-85c8-2658be78bdf0",
		Name:         "Master Splicer",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.TribalAnthem(golems, 1, 1)},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Master Splicer — create a 3/3 colorless Phyrexian Golem artifact creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("3/3 colorless Phyrexian Golem artifact"), N: 1})),
		},
	})
}
