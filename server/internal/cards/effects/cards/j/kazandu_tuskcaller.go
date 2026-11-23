package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kazandu Tuskcaller — Creature — Human Shaman {1}{G}, 1/1 (Rise of the
// Eldrazi, #5581):
//
//	"Level up {1}{G} ({1}{G}: Put a level counter on this. Level up only as
//	 a sorcery.)
//	 LEVEL 2-5
//	 1/1
//	 {T}: Create a 3/3 green Elephant creature token.
//	 LEVEL 6+
//	 1/1
//	 {T}: Create two 3/3 green Elephant creature tokens."
//
// A leveler (leveler.go, CR 702.87a, 711.2a, 711.2b). Each band's tap
// ability is a bundle the Tuskcaller grants itself behind that band's level
// check.
//
// No simplification.
const (
	KazanduTuskcallerOne = "kazandu-tuskcaller/one-elephant"
	KazanduTuskcallerTwo = "kazandu-tuskcaller/two-elephants"
)

func init() {
	effects.Register(effects.Spec{
		OracleID:     "ee60ccae-b3ae-47a1-a0bb-2fe51645e95d",
		Name:         "Kazandu Tuskcaller",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{
			effects.RoeB1ElephantGrant(KazanduTuskcallerOne, 1, "{T}: Create a 3/3 green Elephant creature token."),
			effects.RoeB1ElephantGrant(KazanduTuskcallerTwo, 2, "{T}: Create two 3/3 green Elephant creature tokens."),
		},
		Activated: []effects.ActivatedAbility{effects.LevelUpKeyword("{1}{G}")},
		Static: effects.LevelBands(
			effects.LevelBand{From: 2, To: 5, Power: 1, Toughness: 1, Statics: []game.StaticAbility{
				effects.Labelled(effects.GrantAbilities(effects.SelfOnly, KazanduTuskcallerOne), `This creature has "{T}: Create a 3/3 green Elephant creature token."`),
			}},
			effects.LevelBand{From: 6, Power: 1, Toughness: 1, Statics: []game.StaticAbility{
				effects.Labelled(effects.GrantAbilities(effects.SelfOnly, KazanduTuskcallerTwo), `This creature has "{T}: Create two 3/3 green Elephant creature tokens."`),
			}},
		),
	})
}
