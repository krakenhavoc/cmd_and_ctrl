package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Overabundance — Enchantment {1}{R}{G} (Invasion, S82 #5616):
//
//	"Whenever a player taps a land for mana, that player adds one mana of any type
//	 that land produced, and this enchantment deals 1 damage to the player."
//
// Mana Flare's triggered mana ability (CR 605.1b, no stack) for every player's
// land, with the damage as the ability resolves: after the extra mana is added,
// the enchantment deals 1 damage to the player whose pool it went to
// (game.ManaTrigger.SourceDamagesTapper), the enchantment being the source.
//
// No simplification.
func init() {
	trigger := effects.WheneverAPlayerTapsALandForMana(
		"Overabundance — that player adds one mana of any type that land produced and this enchantment deals 1 damage to them",
		effects.AddsOneManaOfAnyTypeProduced())
	trigger.SourceDamagesTapper = 1
	effects.Register(effects.Spec{
		OracleID:     "b1ee0503-be92-4a06-89cd-5dfa937930d3",
		Name:         "Overabundance",
		Completeness: effects.CompletenessFull,
		ManaTriggers: []game.ManaTrigger{trigger},
	})
}
