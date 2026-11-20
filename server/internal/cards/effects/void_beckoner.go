package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Void Beckoner — Creature — Nightmare Horror {6}{B}{B}, 8/8 (Ikoria: Lair of Behemoths, #5512):
//
//	"Deathtouch
//	 Cycling {2}{B} ({2}{B}, Discard this card: Draw a card.)
//	 When you cycle this card, put a deathtouch counter on target creature you
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
		OracleID:        "7185e345-d8a1-4dd3-a071-f90f265634b4",
		Name:            "Void Beckoner",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Activated:       []ActivatedAbility{Cycling("{2}{B}")},
		Triggered:       []game.TriggeredAbility{ikoCycledCounterOnYourCreature("Void Beckoner", game.CounterDeathtouch, "deathtouch")},
	})
}
