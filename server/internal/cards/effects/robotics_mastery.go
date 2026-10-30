package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Robotics Mastery — Enchantment — Aura {4}{U} (Marvel Super Heroes
// Commander, #4455):
//
//	"Flash
//	 Enchant creature
//	 When this Aura enters, create two 1/1 colorless Robot artifact
//	 creature tokens with flying.
//	 Enchanted creature gets +2/+2."
//
// Flash is a printed keyword, the host gets PumpAttached(2, 2), and the
// enters trigger is the Aura's own, so the Robots arrive whatever
// happens to the host.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "855318cb-9382-4183-87c7-93bdfac86431",
		Name:            "Robotics Mastery",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		Static:          []game.StaticAbility{PumpAttached(2, 2)},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Robotics Mastery — create two 1/1 colorless Robot artifact creature tokens with flying",
				Do(CreateToken{Template: TokenCard("1/1 colorless Robot artifact with flying"), N: 2})),
		},
	})
}
