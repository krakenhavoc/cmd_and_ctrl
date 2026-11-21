package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brotherhood Headquarters — Land:
//
//	"{T}: Add {C}.
//	 {T}: Add one mana of any color. Spend this mana only to cast an
//	 Assassin spell or a spell that has freerunning, or to activate an
//	 ability of an Assassin source."
//
// The restriction is one CR 106.6 tag with three alternatives: a cast of
// an Assassin spell, a cast of a spell that has freerunning (whatever
// cost it is paid with: ManaRestrictSpellFreerunning reads the card's
// offered freerunning cost), or an activation whose source is an
// Assassin.
func init() {
	Register(Spec{
		OracleID:     "0d3a06d5-5bb9-4733-a55b-9e2c75de6b6e",
		Name:         "Brotherhood Headquarters",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			cmdr2B01TapForColor("C"),
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{W|U|B|R|G}",
				Label:    "Add one mana of any color. Spend this mana only to cast an Assassin spell or a spell that has freerunning, or to activate an ability of an Assassin source",
				Restrictions: []string{game.ManaRestrictAnyOf(
					[]string{ManaRestrictCast, ManaRestrictSubtype("Assassin")},
					[]string{game.ManaRestrictSpellFreerunning},
					[]string{ManaRestrictActivate, ManaRestrictSubtype("Assassin")},
				)},
			},
		},
	})
}
