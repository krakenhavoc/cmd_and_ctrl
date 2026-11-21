package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Prickly Marmoset — Creature — Monkey {2}{R}, 2/2 (Jumpstart 2022, #5948):
//
//	"First strike
//	 Whenever you cycle a card, this creature gets +2/+0 until end of turn."
//
// A cycling ability emits EventCycle for its controller (CR 702.29b), so the
// watcher on the battlefield sees it; the Marmoset itself is pumped.
//
// No simplification.
func init() {
	const label = "Prickly Marmoset — it gets +2/+0 until end of turn"
	Register(Spec{
		OracleID:        "a1ef3b31-b996-48ea-967a-2b56fb2fb22b",
		Name:            "Prickly Marmoset",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			WheneverYouCycle(label, func(g *game.Game, item *game.StackItem) error {
				return BoostUntilEOT{Target: item.SourceCardID, Power: 2, Label: label}.Apply(NewContext(g, item))
			}),
		},
	})
}
