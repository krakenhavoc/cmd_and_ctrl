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
// Vigilance and the mana grant are ordinary printed abilities and
// ship complete. The dies-and-return clause is the Enduring cycle's
// shared gap (see Enduring Tenacity, Enduring Curiosity): there is no
// per-instance "lost its creature type permanently" state to hang a
// returned object's stripped Creature type on, and the layer-4
// machinery is keyed on the catalog entry shared by every copy of the
// card. Shipping the return WITHOUT the type change would make this a
// free, unkillable, infinitely recursive 4/3 mana dork — far
// STRONGER than printed — so the clause is dropped rather than
// approximated, same as its cycle-mates. The card as registered dies
// once, which is weaker than printed and therefore the right
// direction.
func init() {
	const grant = "enduring-vitality/any-color"
	Register(Spec{
		OracleID:        "3577c47e-76d3-4659-b922-31c4b74be3a0",
		Name:            "Enduring Vitality",
		Completeness:    CompletenessCaveats,
		PrintedKeywords: []string{"vigilance"},
		Caveats: []string{
			"When this dies, it doesn't return to the battlefield as an enchantment — it goes to the graveyard like any other creature.",
		},
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
	})
}
