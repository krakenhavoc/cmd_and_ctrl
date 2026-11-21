package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	GuulDrazAssassinTwo  = "guul-draz-assassin/minus-two"
	GuulDrazAssassinFour = "guul-draz-assassin/minus-four"
)

func init() {
	effects.Register(effects.Spec{
		OracleID:     "c6fb6971-112b-4c82-a0a3-4acdde970b30",
		Name:         "Guul Draz Assassin",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{
			effects.RoeB1ShrinkGrant(GuulDrazAssassinTwo, "Guul Draz Assassin", 2),
			effects.RoeB1ShrinkGrant(GuulDrazAssassinFour, "Guul Draz Assassin", 4),
		},
		Activated: []effects.ActivatedAbility{effects.LevelUpKeyword("{1}{B}")},
		Static: effects.LevelBands(
			effects.LevelBand{From: 2, To: 3, Power: 2, Toughness: 2, Statics: []game.StaticAbility{
				effects.Labelled(effects.GrantAbilities(effects.SelfOnly, GuulDrazAssassinTwo), `This creature has "{B}, {T}: Target creature gets -2/-2 until end of turn."`),
			}},
			effects.LevelBand{From: 4, Power: 4, Toughness: 4, Statics: []game.StaticAbility{
				effects.Labelled(effects.GrantAbilities(effects.SelfOnly, GuulDrazAssassinFour), `This creature has "{B}, {T}: Target creature gets -4/-4 until end of turn."`),
			}},
		),
	})
}
