package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Creeping Peeper — Creature — Eye {1}{U}, 2/1:
//
//	"{T}: Add {U}. Spend this mana only to cast an enchantment spell,
//	 unlock a door, or turn a permanent face up."
//
// The mana is restricted with ManaRestrictAnyOf: an enchantment spell
// (cast), or an unlock cost (ManaRestrictUnlock, CR 709.5e), or a turn-face-up cost.
//
// Turning a permanent face up (#3172) is the special action's own spend
// purpose (game.SpendPurposeTurnFaceUp, ManaRestrictTurnFaceUp), so the
// mana pays a morph, megamorph or disguise cost, and a manifested or
// cloaked creature's turn-up, and nothing else.
func init() {
	Register(Spec{
		OracleID:     "acf22340-7ff9-443f-a846-c44bb5a3f4e7",
		Name:         "Creeping Peeper",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "Add {U}. Spend this mana only to cast an enchantment spell, unlock a door, or turn a permanent face up",
			Restrictions: []string{game.ManaRestrictAnyOf(
				[]string{game.ManaRestrictCast, game.ManaRestrictType("Enchantment")},
				[]string{game.ManaRestrictUnlock},
				[]string{game.ManaRestrictTurnFaceUp},
			)},
		}},
	})
}
