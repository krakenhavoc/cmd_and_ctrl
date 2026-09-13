package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kalonian Hydra — Creature — Hydra {3}{G}{G}, 0/0 (EDHREC rank 1120):
//
//	"Trample
//	 This creature enters with four +1/+1 counters on it.
//	 Whenever this creature attacks, double the number of +1/+1
//	 counters on each creature you control."
//
// Three mechanisms, all real:
//
//   - Trample rides PrintedKeywords.
//   - "Enters with four +1/+1 counters" is a self-replacement on its
//     own entry (Mossborn Hydra's hook with a count), so the counters
//     are on the 0/0 before any state check sees it.
//   - The attack trigger is Bristly Bill's activated body — for each
//     creature you control, as many +1/+1 counters again as it has —
//     snapshotted before anything is placed, so a counter doubler
//     that fires on the first creature cannot change what the second
//     receives. Placed through AddCounter, so Doubling Season and
//     Hardened Scales apply to the doubling, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7bd36106-04fe-481f-b16e-e076dcbb183b",
		Name:            "Kalonian Hydra",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters("+1/+1", 4, "Kalonian Hydra: enters with four +1/+1 counters"),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclared(ev, source)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Kalonian Hydra — double the +1/+1 counters on each creature you control",
					b08DoubleCountersOnEachCreatureYouControl)
			},
		}},
	})
}
