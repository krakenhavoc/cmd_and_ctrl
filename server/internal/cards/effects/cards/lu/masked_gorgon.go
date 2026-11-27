package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Masked Gorgon — Creature — Gorgon {4}{B}, 5/5 (Judgment, #5611):
//
//	"Green creatures and white creatures have protection from Gorgons.
//	 Threshold — This creature has protection from green and from white as
//	 long as there are seven or more cards in your graveyard."
//
// A layer 6 grant of "protection from Gorgons" to every green or white
// creature (any controller; the Gorgon itself is black), and the threshold
// self-grant of the two colour protections.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f41b1a5b-bc99-4c51-9665-e251f04ad83f",
		Name:         "Masked Gorgon",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.Labelled(effects.B16GrantKeywords(effects.OldB622GreenOrWhiteCreature, "protection from Gorgons"),
				"Green creatures and white creatures have protection from Gorgons."),
			effects.FraThresholdSelfKeywords(game.ProtectionFromColor("G"), game.ProtectionFromColor("W")),
		},
	})
}
