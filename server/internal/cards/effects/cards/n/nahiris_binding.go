package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nahiri's Binding — Enchantment — Aura {1}{W}{W} (Zendikar Rising, #5508):
//
//	"Enchant creature or planeswalker
//	 Enchanted permanent can't attack or block, and its activated abilities
//	 can't be activated."
//
// Planar Disruption's lock on a narrower host. The printed text has no
// mana-ability rider, so CantActivateMana is set as well (CR 605.1a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8f674382-78b5-462c-9a7f-c5f613c317d4",
		Name:         "Nahiri's Binding",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("enchant creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			effects.RestrictAttached(game.CantAttackOrBlock | game.CantActivate | game.CantActivateMana),
		},
	})
}
