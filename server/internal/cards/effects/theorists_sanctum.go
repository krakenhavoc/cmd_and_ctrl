package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Theorist's Sanctum — Land — Island (Reality Fracture, tracker #2795):
//
//	"({T}: Add {U}.)
//	 As this land enters, you may behold a Jace. If you don't, this land
//	 enters tapped. (To behold a Jace, choose a Jace you control or
//	 reveal a Jace card from your hand.)
//	 {2}{U}, {T}: Empower Jace 2."
//
// The entry is the reveal-lands' replacement (reveal_lands.go) with the
// "choose a Jace you control" half of behold folded in: a controlled Jace
// (a planeswalker or the Jace token) is free and always at least as good
// as the tapped alternative, so with one the land simply enters untapped
// without a prompt; otherwise the player is asked whether to reveal a
// Jace card from hand. The mana ability is declared because the engine
// only derives one for basic lands.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "37513bd8-7303-4ac6-a7a7-224184faf758",
		Name:         "Theorist's Sanctum",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			frEntersTappedUnlessYouBeholdAJace("Theorist's Sanctum"),
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "Add {U}",
		}},
		Activated: []ActivatedAbility{{
			Label: "{2}{U}, {T}: Empower Jace 2.",
			Cost:  Plus(ManaCost("{2}{U}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return EmpowerJace{N: 2}.Apply(NewContext(g, item))
			},
		}},
	})
}
