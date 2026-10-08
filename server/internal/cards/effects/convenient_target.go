package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Convenient Target — Enchantment — Aura {R}:
//
//	"Enchant creature
//	 When this Aura enters, suspect enchanted creature. (It has menace
//	 and can't block.)
//	 Enchanted creature gets +1/+1.
//	 {2}{R}: Return this card from your graveyard to your hand."
//
// The enters trigger is suspectEnchantedCreature, which names the
// creature through the Aura's attachment as it resolves, or as it last
// was if the Aura has already gone, so "enchanted creature" still means
// the host. The recursion is an ability of the card in the graveyard
// (Eldrazi Ravager's shape).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2f214dcf-9601-4ccb-bdac-2f9b9f6dff99",
		Name:         "Convenient Target",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static:       []game.StaticAbility{PumpAttached(1, 1)},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Convenient Target — suspect enchanted creature", suspectEnchantedCreature),
		},
		Activated: []ActivatedAbility{{
			Label:  "{2}{R}: Return this card from your graveyard to your hand.",
			Cost:   ManaCost("{2}{R}"),
			Zones:  []game.ZoneKind{game.ZoneGraveyard},
			Effect: returnThisCardFromYourGraveyardToYourHand,
		}},
	})
}
