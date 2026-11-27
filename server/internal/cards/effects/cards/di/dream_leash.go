package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dream Leash — Enchantment — Aura {3}{U}{U} (Ravnica: City of Guilds, S82 OLD
// batch 581, #5600):
//
//	"Enchant permanent
//	 You can't choose an untapped permanent as this spell's target as you
//	 cast it.
//	 You control enchanted permanent."
//
// Enthralling Hold's shape (#6168) over any permanent: the enchant clause
// carries TappedOnlyAsCast, read by the legal set and the announce check of
// the cast and by nothing after it, so the Aura stays attached once the
// permanent untaps (CR 704.5m asks only the enchant ability). The control
// change is the Aura's layer-2 static (CR 613.2b), ending with the Aura.
//
// No simplification.
func init() {
	targets := effects.EnchantPermanent()
	targets.TappedOnlyAsCast = true
	effects.Register(effects.Spec{
		OracleID:     "b7fde102-7b89-48c7-818c-bdf0eb33c2ee",
		Name:         "Dream Leash",
		Completeness: effects.CompletenessFull,
		Targets:      targets,
		Static: []game.StaticAbility{
			effects.ControlAttachedBySource(),
		},
	})
}
