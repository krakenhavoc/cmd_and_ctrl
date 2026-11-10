package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Abundant Countryside — Land:
//
//	"{T}: Add {C}.
//	 {T}: Add one mana of any color. Spend this mana only to cast a
//	 creature spell.
//	 {6}, {T}: Create a 1/1 colorless Shapeshifter creature token with
//	 changeling. (It's every creature type.)"
//
// Two separate mana abilities (the plain {C} and Ancient Ziggurat's
// restricted one) and a token maker with Castle Ardenvale's shape.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e3eb6f90-ccfc-41e7-bff6-0b378226bc7e",
		Name:         "Abundant Countryside",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			painlessColorless(),
			cmdr2B01TapForCreatureOnlyMana(),
		},
		Activated: []ActivatedAbility{{
			Label:   "{6}, {T}: Create a 1/1 colorless Shapeshifter creature token with changeling.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    Plus(ManaCost("{6}"), TapCost()),
			Effect:  cmdr2B01MakeToken("1/1 colorless Shapeshifter with changeling"),
		}},
	})
}
