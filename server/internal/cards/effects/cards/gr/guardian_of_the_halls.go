package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Guardian of the Halls — Creature — Elf Soldier {1}{G}, 2/2 (The Hobbit,
// HOB):
//
//	"Trample
//	 {5}{G}{G}: Put three +1/+1 counters on this creature."
//
// Trample rides PrintedKeywords. The counters go on the permanent that
// was activated, so an activation whose source left the battlefield and
// came back as a new object does nothing (CR 400.7), as Feral Hydra's
// does (#4454).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c3fb2829-64f2-4795-b815-edcc5bbbc360",
		Name:            "Guardian of the Halls",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}{G}{G}: Put three +1/+1 counters on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{5}{G}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 3}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
