package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Strange Augmentation — Enchantment — Aura {B}, (Eldritch Moon, S82 OLD,
// #5541):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1.
//	 Delirium — Enchanted creature gets an additional +2/+2 as long as
//	 there are four or more card types among cards in your graveyard."
//
// Two layer 7c modifies on the host. The delirium one counts the Aura
// controller's own graveyard (not the host controller's), re-read on every
// recompute (CR 207.2c).
//
// No simplification.
func init() {
	bonus := effects.PumpAttachedPer(2, 2, func(g *game.Game, source *game.Card) int {
		if effects.DrcDeliriumMet(g, source.Controller) {
			return 1
		}
		return 0
	})
	bonus.Label = "Delirium — Enchanted creature gets an additional +2/+2 as long as there are four or more card types among cards in your graveyard."
	base := effects.PumpAttached(1, 1)
	base.Label = "Enchanted creature gets +1/+1."
	effects.Register(effects.Spec{
		OracleID:     "5a499546-eab4-40e7-8f7b-4d318aa25810",
		Name:         "Strange Augmentation",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{base, bonus},
	})
}
