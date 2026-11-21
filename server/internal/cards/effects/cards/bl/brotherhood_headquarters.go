package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "0d3a06d5-5bb9-4733-a55b-9e2c75de6b6e",
		Name:         "Brotherhood Headquarters",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			effects.Cmdr2B01TapForColor("C"),
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{W|U|B|R|G}",
				Label:    "Add one mana of any color. Spend this mana only to cast an Assassin spell or a spell that has freerunning, or to activate an ability of an Assassin source",
				Restrictions: []string{game.ManaRestrictAnyOf(
					[]string{effects.ManaRestrictCast, effects.ManaRestrictSubtype("Assassin")},
					[]string{game.ManaRestrictSpellFreerunning},
					[]string{effects.ManaRestrictActivate, effects.ManaRestrictSubtype("Assassin")},
				)},
			},
		},
	})
}
