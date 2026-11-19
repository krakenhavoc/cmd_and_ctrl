package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rescuer Sphinx — Creature — Sphinx {2}{U}{U}, 3/2 (Ravnica: Clue Edition,
// #5929):
//
//	"Flying
//	 As this creature enters, you may return a nonland permanent you control
//	 to its owner's hand. If you do, this creature enters with a +1/+1
//	 counter on it."
//
// CR 614.1c makes "as this enters" a replacement effect, so the choice is
// the EntryCardChoose of ADR 0170 with its Bounce flag: the return happens
// while the entry is paused, and the counter rides the entry (CR 614.12),
// only if the permanent really reached a hand. The Sphinx itself and
// anything entering with it are never candidates (CR 614.13a).
//
// No simplification.
func init() {
	const name = "Rescuer Sphinx"
	Register(Spec{
		OracleID:        "d4c4ad11-7e71-4ad0-aee7-58ac7b1b38f8",
		Name:            name,
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements: []game.ReplacementEffect{
			AsEntersChoosePermanents(name,
				"Rescuer Sphinx — as it enters, you may return a nonland permanent you control to its owner's hand. If you do, it enters with a +1/+1 counter on it.",
				&game.EntryCardChoice{
					Action:  game.EntryCardChoose,
					Matches: func(c game.Card) bool { return !c.IsLand() },
					Min:     0,
					Max:     1,
					Bounce:  true,
					Then:    rescuerSphinxCounter,
				}),
		},
	})
}

// rescuerSphinxCounter is the entry's Then, run only when the permanent
// reached a hand: one +1/+1 counter as the Sphinx enters.
func rescuerSphinxCounter(_ *game.Game, ev *game.ReplacementEvent, picked []uuid.UUID) error {
	if len(picked) > 0 {
		ev.AddCounterAtETB(game.CounterPlusOne, 1)
	}
	return nil
}
