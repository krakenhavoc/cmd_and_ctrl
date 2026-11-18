package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "e0717287-9bff-446b-838d-6fc0adccc768",
		Name:            "Cradle of Safety",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         EnchantCreature(YouControl()),
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Cradle of Safety — enchanted creature gains hexproof until end of turn", acceleratedEvolutionEnters),
		},
	})
}
