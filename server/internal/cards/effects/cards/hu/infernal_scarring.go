package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Infernal Scarring — Enchantment — Aura {1}{B} (Core Set 2021, #5511):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+0 and has "When this creature dies, draw a
//	 card.""
//
// The granted trigger is the creature's own (ADR 0093): its controller as it
// died draws. The bundle is given by the Aura's static while it is attached.
//
// No simplification.
func init() {
	const label = "Infernal Scarring — draw a card"
	const key = "infernal-scarring/dies-draw"
	effects.Register(effects.Spec{
		OracleID:     "3fc640dd-6292-4d36-82fb-bd366bc00bd3",
		Name:         "Infernal Scarring",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Grants: []effects.AbilityGrant{{
			Key:       key,
			Triggered: []game.TriggeredAbility{effects.WhenThisDies(label, effects.Do(effects.DrawCards{N: 1}))},
			Text:      "When this creature dies, draw a card.",
		}},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 0),
			effects.GrantAbilitiesToAttached(key),
		},
	})
}
