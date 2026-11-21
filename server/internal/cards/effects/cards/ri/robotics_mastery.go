package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "855318cb-9382-4183-87c7-93bdfac86431",
		Name:            "Robotics Mastery",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		Static:          []game.StaticAbility{effects.PumpAttached(2, 2)},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Robotics Mastery — create two 1/1 colorless Robot artifact creature tokens with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Robot artifact with flying"), N: 2})),
		},
	})
}
