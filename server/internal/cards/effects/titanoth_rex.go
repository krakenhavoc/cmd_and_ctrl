package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Titanoth Rex — Creature — Dinosaur Beast {7}{G}{G}, 11/11 (Ikoria: Lair of Behemoths, #5512):
//
//	"Trample
//	 Cycling {1}{G} ({1}{G}, Discard this card: Draw a card.)
//	 When you cycle this card, put a trample counter on target creature you
//	 control."
//
// The cycling trigger fires from the graveyard (CR 702.29c) and goes on
// the stack above the cycling ability, so the counter lands before the
// draw. The counter gives the keyword for as long as it stays (CR
// 122.1b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8656a32b-eb95-402b-9daf-0b6d876b4b13",
		Name:            "Titanoth Rex",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated:       []ActivatedAbility{Cycling("{1}{G}")},
		Triggered:       []game.TriggeredAbility{ikoCycledCounterOnYourCreature("Titanoth Rex", game.CounterTrample, "trample")},
	})
}
