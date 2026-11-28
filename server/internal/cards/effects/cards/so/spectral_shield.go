package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spectral Shield — Enchantment — Aura {1}{W}{U} (Ice Age, S82 OLD b828,
// #5636):
//
//	"Enchant creature
//	 Enchanted creature gets +0/+2 and can't be the target of spells."
//
// PumpAttached and the attached spell-targeting bar (AttachedCantBeTargetedBySpells);
// abilities still target the creature, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "37d9c065-598d-4ee3-abfc-625bd6169ff1",
		Name:         "Spectral Shield",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(0, 2),
			effects.AttachedCantBeTargetedBySpells("Enchanted creature can't be the target of spells."),
		},
	})
}
