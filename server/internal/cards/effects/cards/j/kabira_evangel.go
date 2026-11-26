package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kabira Evangel — Creature — Human Cleric Ally {2}{W}, 2/3 (Zendikar, S82 OLD
// b494):
//
//	"Whenever this creature or another Ally you control enters, you may choose a
//	 color. If you do, Allies you control gain protection from the chosen color
//	 until end of turn."
//
// A rally trigger (one per entering Ally), optional, whose "if you do" asks for
// the colour and hands the Allies read as it resolves to Bathe in Light's
// protection tail (RadianceProtectionFromChosenColor).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "83b480ea-77c5-4e00-99d0-bea2a60a1bfe",
		Name:         "Kabira Evangel",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.S82OldB313Rally("Kabira Evangel — choose a color; Allies you control gain protection from it until end of turn", effects.Old494AlliesGainProtection),
				"Kabira Evangel — choose a color for your Allies to gain protection from?"),
		},
	})
}
