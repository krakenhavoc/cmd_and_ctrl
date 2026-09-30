package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Creeping Peeper — Creature — Eye {1}{U}, 2/1:
//
//	"{T}: Add {U}. Spend this mana only to cast an enchantment spell,
//	 unlock a door, or turn a permanent face up."
//
// The mana is restricted with ManaRestrictAnyOf: an enchantment spell
// (cast), or an unlock cost (ManaRestrictUnlock, CR 709.5e).
//
// One declared gap, weaker than printed: the engine has no spend
// purpose for turning a permanent face up (morph, disguise, manifest
// and cloak costs), so this mana cannot pay those. A restriction the
// matcher does not understand would deny, never allow, so the gap only
// ever costs the player, never favours them.
func init() {
	Register(Spec{
		OracleID:     "acf22340-7ff9-443f-a846-c44bb5a3f4e7",
		Name:         "Creeping Peeper",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The mana can cast enchantment spells and unlock doors, but it can't be spent to turn a permanent face up.",
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "Add {U}. Spend this mana only to cast an enchantment spell or unlock a door",
			Restrictions: []string{game.ManaRestrictAnyOf(
				[]string{game.ManaRestrictCast, game.ManaRestrictType("Enchantment")},
				[]string{game.ManaRestrictUnlock},
			)},
		}},
	})
}
