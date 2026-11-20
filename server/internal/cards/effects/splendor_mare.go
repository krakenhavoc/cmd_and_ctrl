package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Splendor Mare — Creature — Elk Unicorn {2}{W}, 3/3 (Ikoria: Lair of Behemoths, #5512):
//
//	"Lifelink
//	 Cycling {1}{W} ({1}{W}, Discard this card: Draw a card.)
//	 When you cycle this card, put a lifelink counter on target creature you
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
		OracleID:        "fd2a00d1-9f21-4d54-9458-024fa03354b8",
		Name:            "Splendor Mare",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Activated:       []ActivatedAbility{Cycling("{1}{W}")},
		Triggered:       []game.TriggeredAbility{ikoCycledCounterOnYourCreature("Splendor Mare", game.CounterLifelink, "lifelink")},
	})
}
