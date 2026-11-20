package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flinthoof Boar — Creature — Boar {1}{G}, 2/2 (Magic 2013, #5568):
//
//	"This creature gets +1/+1 as long as you control a Mountain.
//	 {R}: This creature gains haste until end of turn. (It can attack and
//	 {T} this turn.)"
//
// Kird Ape's land-type pump and Thornling's haste for mana.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "acb58839-2b26-4724-9e9b-b49b8217b13f",
		Name:         "Flinthoof Boar",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "This creature gets +1/+1 as long as you control a Mountain.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: selfWhileYouControlA(Subtype("Mountain")),
			Apply:     pumpBy(1, 1),
		}},
		Activated: []ActivatedAbility{conB1ThisGainsRow("{R}", "haste", "Flinthoof Boar", game.AnswerCombatGrant)},
	})
}
