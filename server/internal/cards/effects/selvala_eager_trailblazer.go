package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Selvala, Eager Trailblazer — Legendary Creature — Elf Scout {2}{G}{W},
// 4/5 (Outlaws of Thunder Junction, #4468):
//
//	"Vigilance
//	 Whenever you cast a creature spell, create a 1/1 red Mercenary
//	 creature token with '{T}: Target creature you control gets +1/+0
//	 until end of turn. Activate only as a sorcery.'
//	 {T}: Choose a color. Add one mana of that color for each different
//	 power among creatures you control."
//
// The cast trigger is Beast Whisperer's (YouCastACreatureSpell) making
// the set's Mercenary (MercenaryToken). The mana ability is Loot, the
// Nexus's: one colour pick, then one mana of it for each distinct power
// value among the creatures you control as it is activated. Selvala is a
// creature with a tap cost, so summoning sickness applies.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "95b20721-1af9-4f3f-a39b-7cd1d4ce0b20",
		Name:            "Selvala, Eager Trailblazer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, YouCastACreatureSpell,
				"Selvala, Eager Trailblazer — create a 1/1 red Mercenary", createAMercenary()),
		},
		ManaAbilities: []ManaAbility{{
			Cost:           ManaAbilityCost{Tap: true},
			ProducedAmount: ProducedOneColor(rfCreatureCDifferentPowers),
			Label:          "{T}: Choose a color. Add one mana of that color for each different power among creatures you control",
		}},
	})
}
