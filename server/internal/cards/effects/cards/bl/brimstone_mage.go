package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brimstone Mage — Creature — Human Shaman {2}{R}, 2/2 (Rise of the Eldrazi,
// #5581):
//
//	"Level up {3}{R} ({3}{R}: Put a level counter on this. Level up only as
//	 a sorcery.)
//	 LEVEL 1-2
//	 2/3
//	 {T}: This creature deals 1 damage to any target.
//	 LEVEL 3+
//	 2/4
//	 {T}: This creature deals 3 damage to any target."
//
// A leveler (leveler.go, CR 702.87a, 711.2a, 711.2b). Each band's ping is a
// bundle the Mage grants itself behind that band's level check.
//
// No simplification.
const (
	brimstoneMageOne   = "brimstone-mage/one-damage"
	BrimstoneMageThree = "brimstone-mage/three-damage"
)

func init() {
	effects.Register(effects.Spec{
		OracleID:     "81dcb85e-d61f-430b-9540-5a3bb1378d5a",
		Name:         "Brimstone Mage",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{
			effects.RoeB1PingGrant(brimstoneMageOne, 1),
			effects.RoeB1PingGrant(BrimstoneMageThree, 3),
		},
		Activated: []effects.ActivatedAbility{effects.LevelUpKeyword("{3}{R}")},
		Static: effects.LevelBands(
			effects.LevelBand{From: 1, To: 2, Power: 2, Toughness: 3, Statics: []game.StaticAbility{
				effects.Labelled(effects.GrantAbilities(effects.SelfOnly, brimstoneMageOne), `This creature has "{T}: This creature deals 1 damage to any target."`),
			}},
			effects.LevelBand{From: 3, Power: 2, Toughness: 4, Statics: []game.StaticAbility{
				effects.Labelled(effects.GrantAbilities(effects.SelfOnly, BrimstoneMageThree), `This creature has "{T}: This creature deals 3 damage to any target."`),
			}},
		),
	})
}
