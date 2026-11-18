package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dawnhart Rejuvenator — Creature — Human Warlock {3}{G}, 2/4 (Innistrad: Midnight Hunt, S82):
//
//	"When this creature enters, you gain 3 life.
//	 {T}: Add one mana of any color."
//
// The life gain is a stack trigger to the controller; the mana ability is the usual tap-for-any-colour shape.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c6e92911-a100-43eb-ab4b-3778b0db1918",
		Name:         "Dawnhart Rejuvenator",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "{T}: Add one mana of any color.",
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Dawnhart Rejuvenator — you gain 3 life", Do(GainLife{Amount: 3})),
		},
	})
}
