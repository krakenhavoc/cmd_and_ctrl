package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shang-Chi, Master of Kung Fu — {1}{G} Legendary Creature — Human Warrior
// Hero, 2/2 (Marvel Super Heroes, #4455):
//
//	"You may activate abilities of creatures you control as though those
//	 creatures had haste.
//	 {T}: Add two mana of any one color. Spend this mana only to activate
//	 abilities of creature sources."
//
// The first line is Thousand-Year Elixir's ActivateAsThoughHaste (CR 302.6
// stops paying a {T} cost the turn a creature arrives; this "as though"
// effect lifts it for Shang-Chi's own mana ability too). The mana is
// Gwenna's restriction with only the activation half: the spend tag reads
// the source's types (CR 106.6), and the two mana are one color.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:              "700ef9c8-698b-45f7-a1fd-dec74e553224",
		Name:                  "Shang-Chi, Master of Kung Fu",
		Completeness:          CompletenessFull,
		ActivateAsThoughHaste: true,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: OneColorOfAmount(2),
			Label:    "{T}: Add two mana of any one color. Spend this mana only to activate abilities of creature sources.",
			Restrictions: []string{game.ManaRestrictAnyOf(
				[]string{game.ManaRestrictActivate, game.ManaRestrictType("Creature")},
			)},
		}},
	})
}
