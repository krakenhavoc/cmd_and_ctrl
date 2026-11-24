package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blacker Lotus — Artifact {0} (Secret Lair Drop):
//
//	"{T}: Tear this artifact into pieces. Add four mana of any one color.
//	 Remove the pieces from the game."
//
// A mana ability (CR 605.1a): no target, and it adds mana. "Tear this artifact
// into pieces … remove the pieces from the game" is exiling it, which nobody
// can respond to, so it is paid with the tap, as Ether's "{T}, Exile this
// artifact: Add {U}" is. The four mana are all one colour chosen as it is added.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "41dd29b9-f08d-4ccc-8dc0-da11d2d456e9",
		Name:         "Blacker Lotus",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Zones:    []game.ZoneKind{game.ZoneBattlefield},
			Cost:     effects.ManaAbilityCost{Tap: true, ExileSelf: true},
			Produced: effects.OneColorOfAmount(4),
			Label:    "{T}: Tear this artifact into pieces. Add four mana of any one color. Remove the pieces from the game.",
		}},
	})
}
