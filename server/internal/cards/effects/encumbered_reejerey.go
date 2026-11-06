package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Encumbered Reejerey — Creature — Merfolk Soldier {1}{W}, 5/4 (Lorwyn
// Eclipsed, #4458):
//
//	"This creature enters with three -1/-1 counters on it.
//	 Whenever this creature becomes tapped while it has a -1/-1 counter on
//	 it, remove a -1/-1 counter from it."
//
// Bristlebane Battler's shape with a different event: it enters as a 2/1
// and grows by one each time it becomes tapped, up to 5/4. "Becomes tapped"
// is ThisBecameTapped over the tap event and the attack that taps it, so a
// vigilance attack is not one. "While it has a -1/-1 counter on it" is part
// of the trigger event, read as it becomes tapped, so with no counter left
// it does not trigger; the trigger then removes one if it still has one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7668f7bc-9a43-4024-87c3-995aa0a65aaa",
		Name:         "Encumbered Reejerey",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterMinusOne, 3, "Encumbered Reejerey: enters with three -1/-1 counters"),
		},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, AllOf(ThisBecameTapped, thisHasAMinusOneCounter),
				"Encumbered Reejerey — remove a -1/-1 counter from it",
				removeAMinusOneCounterFromThis),
		},
	})
}
