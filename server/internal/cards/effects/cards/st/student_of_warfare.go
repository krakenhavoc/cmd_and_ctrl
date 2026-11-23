package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Student of Warfare — Creature — Human Knight {W}, 1/1 (Rise of the
// Eldrazi, #5581):
//
//	"Level up {W} ({W}: Put a level counter on this. Level up only as a
//	 sorcery.)
//	 LEVEL 2-6
//	 3/3
//	 First strike
//	 LEVEL 7+
//	 4/4
//	 Double strike"
//
// A leveler (leveler.go, CR 702.87a, 711.2a, 711.2b). With no or one level
// counter it is the printed 1/1 (CR 711.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0bc2875b-3a5a-4885-86b9-16ab7330ac10",
		Name:         "Student of Warfare",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.LevelUpKeyword("{W}")},
		Static: effects.LevelBands(
			effects.LevelBand{From: 2, To: 6, Power: 3, Toughness: 3, Keywords: []string{"first strike"}},
			effects.LevelBand{From: 7, Power: 4, Toughness: 4, Keywords: []string{"double strike"}},
		),
	})
}
