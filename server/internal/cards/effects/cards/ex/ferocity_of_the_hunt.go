package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ferocity of the Hunt — Enchantment — Aura {1}{B/G} (Reality Fracture):
//
//	"Flash
//	 Enchant creature
//	 Enchanted creature gets +1/+0 and has deathtouch.
//	 When enchanted creature dies, return that card to the battlefield
//	 tapped under its owner's control."
//
// The dies trigger is Gift of Immortality's first half: the creature
// card comes back as a new object, tapped, under its OWNER's control.
// A token has no card to return, so it does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "747a1b36-eb4f-44e6-9e5a-bb1a0259a471",
		Name:            "Ferocity of the Hunt",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.EnchantCreature(),
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 0),
			effects.GrantToAttached("deathtouch"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenEnchantedCreatureDies("Ferocity of the Hunt — return that card to the battlefield tapped", effects.ReturnDiedCardToBattlefieldTapped),
		},
	})
}
