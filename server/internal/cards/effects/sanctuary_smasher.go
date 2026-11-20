package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sanctuary Smasher — Creature — Rhino Beast {4}{R}{R}, 6/4 (Ikoria: Lair of Behemoths, #5512):
//
//	"First strike
//	 Cycling {2}{R} ({2}{R}, Discard this card: Draw a card.)
//	 When you cycle this card, put a first strike counter on target creature you
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
		OracleID:        "29925297-b6d4-4fc8-a585-1425c1482ec6",
		Name:            "Sanctuary Smasher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Activated:       []ActivatedAbility{Cycling("{2}{R}")},
		Triggered:       []game.TriggeredAbility{ikoCycledCounterOnYourCreature("Sanctuary Smasher", game.CounterFirstStrike, "first strike")},
	})
}
