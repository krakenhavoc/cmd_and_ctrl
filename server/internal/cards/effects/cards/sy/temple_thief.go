package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Temple Thief — Creature — Human Rogue {1}{B}, 2/2 (Theros Beyond Death,
// #5514):
//
//	"This creature can't be blocked by enchanted creatures or enchantment
//	 creatures."
//
// Beloved Princess's block rule over the blocker. A blocker is refused when an
// Aura is attached to it (an Equipment does not enchant) or when it is an
// enchantment creature, both read live as the block is declared (CR 509.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "493ecd11-e980-42da-a79e-cddd9a94cc7a",
		Name:         "Temple Thief",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(),
				effects.Or(effects.Not(effects.B15NotEnchanted()), effects.Enchantment()),
				"enchanted creatures or enchantment creatures"),
		},
	})
}
