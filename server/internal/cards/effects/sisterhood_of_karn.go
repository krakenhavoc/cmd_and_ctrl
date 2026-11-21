package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sisterhood of Karn — Creature — Cleric {1}{G}, 0/0 (Doctor Who):
//
//	"This creature enters with a +1/+1 counter on it.
//	 Paradox — Whenever you cast a spell from anywhere other than your hand,
//	 double the number of +1/+1 counters on this creature."
//
// The first line is a CR 614.1c self-replacement. The doubling puts as many
// more +1/+1 counters on it as it has as the ability resolves, so a counter
// doubler on the board applies to the placement as to any other; with none
// (or once it has left the battlefield) nothing happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b32484ee-c7d8-472d-818f-2c37783d028a",
		Name:         "Sisterhood of Karn",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters("+1/+1", 1, "This creature enters with a +1/+1 counter on it."),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, gapB85YouCastFromOutsideYourHand,
				"Paradox — double the number of +1/+1 counters on this creature", gapB85DoublePlusOneCountersOnThis),
		},
	})
}
