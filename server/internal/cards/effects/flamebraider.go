package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flamebraider — Creature — Elemental Bard {1}{R}, 2/2 (Lorwyn Eclipsed,
// #4458):
//
//	"{T}: Add two mana in any combination of colors. Spend this mana only
//	 to cast Elemental spells or activate abilities of Elemental sources."
//
// A mana ability (CR 605.1a) producing two mana, each of a color its
// controller names (AnyCombinationOfColors), with Avengers Tower's
// restriction over Elementals: the mana pays for an Elemental spell, or
// for an activated ability of an Elemental source (a permanent, or a card
// in another zone such as a landcycling Elemental in hand, the ruling),
// and for nothing else. A changeling is an Elemental.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c552bd73-ce75-40fd-9d80-9e90b6f45274",
		Name:         "Flamebraider",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: AnyCombinationOfColors(2),
			Label:    "{T}: Add two mana in any combination of colors. Spend this mana only to cast Elemental spells or activate abilities of Elemental sources.",
			Restrictions: []string{game.ManaRestrictAnyOf(
				[]string{game.ManaRestrictCast, game.ManaRestrictSubtype("Elemental")},
				[]string{game.ManaRestrictActivate, game.ManaRestrictSubtype("Elemental")},
			)},
		}},
	})
}
