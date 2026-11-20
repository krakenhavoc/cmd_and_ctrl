package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Guul Draz Assassin — Creature — Vampire Assassin {B}, 1/1 (Rise of the
// Eldrazi, #5581):
//
//	"Level up {1}{B} ({1}{B}: Put a level counter on this. Level up only as
//	 a sorcery.)
//	 LEVEL 2-3
//	 2/2
//	 {B}, {T}: Target creature gets -2/-2 until end of turn.
//	 LEVEL 4+
//	 4/4
//	 {B}, {T}: Target creature gets -4/-4 until end of turn."
//
// A leveler (leveler.go, CR 702.87a, 711.2a, 711.2b). Each band's activated
// ability is a bundle the Assassin grants itself behind that band's level
// check, so it has at most one of them, and none below level 2.
//
// No simplification.
const (
	guulDrazAssassinTwo  = "guul-draz-assassin/minus-two"
	guulDrazAssassinFour = "guul-draz-assassin/minus-four"
)

func init() {
	Register(Spec{
		OracleID:     "c6fb6971-112b-4c82-a0a3-4acdde970b30",
		Name:         "Guul Draz Assassin",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{
			roeB1ShrinkGrant(guulDrazAssassinTwo, "Guul Draz Assassin", 2),
			roeB1ShrinkGrant(guulDrazAssassinFour, "Guul Draz Assassin", 4),
		},
		Activated: []ActivatedAbility{LevelUpKeyword("{1}{B}")},
		Static: LevelBands(
			LevelBand{From: 2, To: 3, Power: 2, Toughness: 2, Statics: []game.StaticAbility{
				Labelled(GrantAbilities(selfOnly, guulDrazAssassinTwo), `This creature has "{B}, {T}: Target creature gets -2/-2 until end of turn."`),
			}},
			LevelBand{From: 4, Power: 4, Toughness: 4, Statics: []game.StaticAbility{
				Labelled(GrantAbilities(selfOnly, guulDrazAssassinFour), `This creature has "{B}, {T}: Target creature gets -4/-4 until end of turn."`),
			}},
		),
	})
}
