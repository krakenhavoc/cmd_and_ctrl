package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dryad's Favor — Enchantment — Aura {G} (Magic 2011, #5579):
//
//	"Enchant creature
//	 Enchanted creature has forestwalk."
//
// Fishliver Oil's grant with forestwalk: an engine keyword the block
// check reads, granted to the enchanted creature in layer 6.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "af740170-af87-487c-abab-a6ccdea82d97",
		Name:         "Dryad's Favor",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static:       []game.StaticAbility{effects.GrantToAttached("forestwalk")},
	})
}
