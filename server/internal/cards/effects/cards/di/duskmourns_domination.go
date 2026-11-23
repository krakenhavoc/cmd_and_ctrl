package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Duskmourn's Domination — Enchantment — Aura {4}{U}{U} (Duskmourn):
//
//	"Enchant creature
//	 You control enchanted creature.
//	 Enchanted creature gets -3/-0 and loses all abilities."
//
// Mind Control's layer-2 control, Aether Meltdown's pump (here -3/-0) and
// Eaten by Piranhas' layer-6 "loses all abilities". Each static carries its
// own timestamp (CR 613.7), so an ability the creature gains AFTER the Aura
// became attached is kept, as the ruling says. Control reverts on its own
// when the Aura leaves.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ebdace8f-8e8e-411b-b1d5-5b27cca7d9fb",
		Name:         "Duskmourn's Domination",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static: []game.StaticAbility{
			effects.ControlAttachedBySource(),
			effects.PumpEnchantedCreature(-3, 0),
			effects.LoseAllAbilities(),
		},
	})
}
