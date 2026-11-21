package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fungal Fortitude — Enchantment — Aura {1}{B} (The Lost Caverns of
// Ixalan, #4470):
//
//	"Flash
//	 Enchant creature
//	 Enchanted creature gets +2/+0.
//	 When enchanted creature dies, return it to the battlefield tapped
//	 under its owner's control."
//
// Ferocity of the Hunt without the deathtouch: a layer 7c +2/+0 on the
// host, and a dies trigger (returnDiedCardToBattlefieldTapped) that
// brings the card back as a new object, tapped, under its OWNER's
// control. A token has no card to return, so it does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cf73d2e7-584f-41d0-8fc7-50d886dcf9fe",
		Name:            "Fungal Fortitude",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.EnchantCreature(),
		Static:          []game.StaticAbility{effects.PumpAttached(2, 0)},
		Triggered: []game.TriggeredAbility{
			effects.WhenEnchantedCreatureDies("Fungal Fortitude — return it to the battlefield tapped", effects.ReturnDiedCardToBattlefieldTapped),
		},
	})
}
