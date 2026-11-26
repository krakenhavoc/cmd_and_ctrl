package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spy Kit — Artifact — Equipment {2} (Conspiracy: Take the Crown, S82 older
// sets batch 273, #5540):
//
//	"Equipped creature gets +1/+1 and has all names of nonlegendary creature
//	 cards in addition to its name.
//	 Equip {2}"
//
// The +1/+1 is PumpAttached. The names are #6257's seam, a layer 3 flag
// (Characteristic.AllNonlegendaryCreatureNames, CR 612, 201.2a): the equipped
// creature has every nonlegendary creature card's name in addition to its
// own, so a Pithing Needle naming one reaches it and two legendary
// permanents it makes share a name for the legend rule. Which names those
// are is the card database's answer, wired by the server at boot
// (game.NonlegendaryCreatureCardName).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4d94f5bd-42c6-4d15-a693-8be189ec3746",
		Name:         "Spy Kit",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.EquippedCreatureIs(game.Layer3Text,
				"Equipped creature has all names of nonlegendary creature cards in addition to its name.",
				func(c *game.Characteristic) { c.AllNonlegendaryCreatureNames = true }),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
