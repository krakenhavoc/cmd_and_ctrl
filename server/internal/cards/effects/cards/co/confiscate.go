package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Confiscate — Enchantment — Aura {4}{U}{U} (Foundations, #4465):
//
//	"Enchant permanent
//	 You control enchanted permanent."
//
// Control Magic's layer-2 control change on the widest enchant clause
// (Faith's Fetters' EnchantPermanent), so a land, an artifact or a
// planeswalker can be taken, and the Aura stays on a host that stops
// being a creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d599a38c-7719-443e-bb59-55bd43a8fee6",
		Name:         "Confiscate",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPermanent(),
		Static: []game.StaticAbility{
			effects.ControlAttachedBySource(),
		},
	})
}
