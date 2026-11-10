package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Coralhelm Commander — Creature — Merfolk Soldier {U}{U}, 2/2 (The
// Lost Caverns of Ixalan Commander, #5426):
//
//	"Level up {1} ({1}: Put a level counter on this. Level up only as a
//	 sorcery.)
//	 LEVEL 2-3
//	 3/3
//	 Flying
//	 LEVEL 4+
//	 4/4
//	 Flying
//	 Other Merfolk creatures you control get +1/+1."
//
// A leveler card (leveler.go, CR 702.87 and CR 711). With no or one
// level counter it is the printed 2/2 with no abilities but level up
// (CR 711.5). With two or three it has base power and toughness 3/3 and
// flying, and with four or more 4/4, flying and the Merfolk anthem
// (CR 711.2a, 711.2b). Level up stays available at every level
// (CR 711.4), though a fifth counter changes nothing.
//
// The anthem is Merfolk Sovereign's: OTHER Merfolk CREATURES YOU
// CONTROL, read by effective subtype, so the Commander never pumps
// itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0ca0a186-026d-4cec-9c23-15d94e6acca9",
		Name:         "Coralhelm Commander",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{LevelUpKeyword("{1}")},
		Static: LevelBands(
			LevelBand{From: 2, To: 3, Power: 3, Toughness: 3, Keywords: []string{"flying"}},
			LevelBand{
				From: 4, Power: 4, Toughness: 4, Keywords: []string{"flying"},
				Statics: []game.StaticAbility{
					TribalAnthem(TribeFilter{Tribes: []string{"Merfolk"}, Others: true, YoursOnly: true}, 1, 1),
				},
			},
		),
	})
}
