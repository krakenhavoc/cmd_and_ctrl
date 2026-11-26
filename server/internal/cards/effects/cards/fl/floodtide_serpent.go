package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Floodtide Serpent — {4}{U} Creature — Serpent 4/4 (Born of the Gods):
//
//	"This creature can't attack unless you return an enchantment you control to its owner's
//	 hand. (This cost is paid as attackers are declared.)"
//
// The #6190 seam: a cost paid in permanents as attackers are declared (CR 508.1h), the
// enchantment named by the player and returned to its owner's hand.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1031bc78-7eca-44aa-9ea8-47fa67b4ce62",
		Name:         "Floodtide Serpent",
		Completeness: effects.CompletenessFull,
		AttackTaxes: []game.AttackTax{
			effects.CantAttackUnlessYouReturn("an enchantment you control", effects.Enchantment()),
		},
	})
}
