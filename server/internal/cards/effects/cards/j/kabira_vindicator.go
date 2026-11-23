package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kabira Vindicator — Creature — Human Knight {3}{W}, 2/4 (Rise of the
// Eldrazi, #5581):
//
//	"Level up {2}{W} ({2}{W}: Put a level counter on this. Level up only as
//	 a sorcery.)
//	 LEVEL 2-4
//	 3/6
//	 Other creatures you control get +1/+1.
//	 LEVEL 5+
//	 4/8
//	 Other creatures you control get +2/+2."
//
// A leveler (leveler.go, CR 702.87a, 711.2a, 711.2b). Each band's anthem is
// the ordinary "other creatures you control" static gated on the band, so
// the two never stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c43c76d5-aa11-4fce-911a-ec60cb06ce2a",
		Name:         "Kabira Vindicator",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.LevelUpKeyword("{2}{W}")},
		Static: effects.LevelBands(
			effects.LevelBand{From: 2, To: 4, Power: 3, Toughness: 6, Statics: []game.StaticAbility{
				effects.Labelled(effects.B16Anthem(effects.OtherCreaturesYouControl, 1, 1), "Other creatures you control get +1/+1."),
			}},
			effects.LevelBand{From: 5, Power: 4, Toughness: 8, Statics: []game.StaticAbility{
				effects.Labelled(effects.B16Anthem(effects.OtherCreaturesYouControl, 2, 2), "Other creatures you control get +2/+2."),
			}},
		),
	})
}
