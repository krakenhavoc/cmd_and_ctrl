package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Troyan, Gutsy Explorer — Legendary Creature — Vedalken Scout {1}{G}{U},
// 1/3 (Secrets of Strixhaven Commander, #4834):
//
//	"{T}: Add {G}{U}. Spend this mana only to cast spells with mana value 5
//	 or greater or spells with {X} in their mana costs.
//	 {U}, {T}: Draw a card, then discard a card."
//
// The mana carries one restriction tag that is an OR (CR 106.6): a cast
// whose spell has mana value 5 or greater (ManaRestrictManaValueAtLeast, the
// value on the stack, {X} at the announced X: CR 202.3e) or a cast whose
// spell has {X} in its printed mana cost (ManaRestrictSpellXCost, which asks
// about the spell's own mana cost, not the cost being paid, and refuses an
// activated ability). Neither pays for an activation, so the mana can't pay
// for the loot or for Troyan's own abilities. The loot is Merfolk Looter's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1135ebc3-9537-4341-956b-6e22d8be5f20",
		Name:         "Troyan, Gutsy Explorer",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}{U}",
			Label:    "{T}: Add {G}{U}. Spend this mana only to cast spells with mana value 5 or greater or spells with {X} in their mana costs.",
			Restrictions: []string{game.ManaRestrictAnyOf(
				[]string{game.ManaRestrictCast, game.ManaRestrictManaValueAtLeast(5)},
				[]string{game.ManaRestrictSpellXCost},
			)},
		}},
		Activated: []ActivatedAbility{{
			Label:   "{U}, {T}: Draw a card, then discard a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{U}"), TapCost()),
			Effect:  func(g *game.Game, item *game.StackItem) error { return lootOne(g, item, 1) },
		}},
	})
}
