package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Enduring Vitality — Enchantment Creature — Elk Glimmer {1}{G}{G},
// 4/3:
//
//	"Vigilance
//	 Creatures you control have '{T}: Add one mana of any color.'
//	 When Enduring Vitality dies, if it was a creature, return it to
//	 the battlefield under its owner's control. It's an enchantment.
//	 (It's not a creature.)"
//
// Vigilance and the mana grant are ordinary printed abilities. The
// dies-and-return clause is the Enduring cycle's shared trigger
// (WhenThisDiesReturnItAsAnEnchantment, glimmer_return.go): the card
// comes back once as an enchantment that is not a creature, and a
// second death finds "if it was a creature" false. The returned
// enchantment is no longer a creature, so it stops granting the mana
// ability to itself, as printed.
//
// No simplification.
func init() {
	const grant = "enduring-vitality/any-color"
	Register(Spec{
		OracleID:        "3577c47e-76d3-4659-b922-31c4b74be3a0",
		Name:            "Enduring Vitality",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Grants: []AbilityGrant{{
			Key:  grant,
			Text: "{T}: Add one mana of any color.",
			Mana: []ManaAbility{{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{W|U|B|R|G}",
				Label:    "Add one mana of any color",
			}},
		}},
		Static: []game.StaticAbility{
			GrantAbilities(b16CreaturesYouControl, grant),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisDiesReturnItAsAnEnchantment("Enduring Vitality"),
		},
	})
}
