package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Enthralling Hold — Enchantment — Aura {3}{U}{U} (Core Set 2021):
//
//	"Enchant creature
//	 You can't choose an untapped creature as this spell's target as you
//	 cast it.
//	 You control enchanted creature."
//
// Mind Control with a cast-time restriction (#6168): the enchant clause
// carries TappedOnlyAsCast, which the legal set and the announce check
// of the cast read and nothing after it does, so once cast it does not
// matter whether the creature is tapped: the spell still resolves if the
// creature untaps in response, and the Aura stays attached (CR 704.5m
// asks only the enchant ability).
func init() {
	targets := effects.EnchantCreature()
	targets.TappedOnlyAsCast = true
	effects.Register(effects.Spec{
		OracleID:     "2f94cb83-1f48-4293-b219-67830c074aeb",
		Name:         "Enthralling Hold",
		Completeness: effects.CompletenessFull,
		Targets:      targets,
		Static: []game.StaticAbility{
			effects.ControlAttachedBySource(),
		},
	})
}
