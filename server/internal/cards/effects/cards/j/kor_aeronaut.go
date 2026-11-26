package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kor Aeronaut — Creature — Kor Soldier {W}{W}, 2/2 (Duel Decks: Elspeth vs.
// Tezzeret, #5578):
//
//	"Kicker {1}{W}
//	 Flying
//	 When this creature enters, if it was kicked, target creature gains flying
//	 until end of turn."
//
// Kicker is an optional additional cost (CR 702.33); the enters trigger has an
// intervening "if it was kicked" (CR 603.4, ThisKickedAtLeast) read off the
// record the resolution carried onto the permanent, so a Kor Aeronaut that
// was not cast is not kicked. The target is any creature, chosen as the
// trigger goes on the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1adb733c-d18a-4019-9229-0ad701a15b1a",
		Name:            "Kor Aeronaut",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		OptionalCosts:   []game.AdditionalCost{effects.Kicker("{1}{W}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedAtLeast(1)),
					"Kor Aeronaut — kicked, target creature gains flying until end of turn",
					effects.S82OldB474TargetGainsFlyingUntilEndOfTurn("Kor Aeronaut")),
				effects.TargetCreature("target creature")),
		},
	})
}
