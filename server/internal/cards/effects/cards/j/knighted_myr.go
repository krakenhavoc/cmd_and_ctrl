package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Knighted Myr — Artifact Creature — Myr Knight {2}{W}, 2/2 (Commander
// Masters, #5948):
//
//	"{2}{W}: Adapt 1. (If this creature has no +1/+1 counters on it, put
//	 a +1/+1 counter on it.)
//	 Whenever one or more +1/+1 counters are put on this creature, it
//	 gains double strike until end of turn."
//
// The trigger fires on any +1/+1 counters placed on it (Adapt, a
// proliferate, another card), once per placement, and gives double
// strike until end of turn (CR 702.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dfee20a6-7737-49a8-b1cb-0a4914ac0ad3",
		Name:         "Knighted Myr",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}: Adapt 1",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{2}{W}"),
			Effect:  func(g *game.Game, item *game.StackItem) error { return effects.AdaptN(g, item, 1) },
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCounterPlaced, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B11CountersWerePlaced(ev, source.InstanceID, "+1/+1", g)
			}, "Knighted Myr — it gains double strike until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"double strike"},
						Label: "Knighted Myr — double strike"}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
