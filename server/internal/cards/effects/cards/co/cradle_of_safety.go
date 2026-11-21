package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cradle of Safety — Enchantment — Aura {1}{U} (Innistrad: Crimson Vow,
// #5495):
//
//	"Flash
//	 Enchant creature you control
//	 When this Aura enters, enchanted creature gains hexproof until end of
//	 turn. (It can't be the target of spells or abilities your opponents
//	 control.)
//	 Enchanted creature gets +1/+1."
//
// Accelerated Evolution's shape with +1/+1 in place of +2/+2: a flash Aura
// on a creature you control, a static on the host, and an enters trigger
// that gives the host hexproof until end of turn as it resolves
// (acceleratedEvolutionEnters, the same sentence).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e0717287-9bff-446b-838d-6fc0adccc768",
		Name:            "Cradle of Safety",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.EnchantCreature(effects.YouControl()),
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Cradle of Safety — enchanted creature gains hexproof until end of turn", effects.AcceleratedEvolutionEnters),
		},
	})
}
