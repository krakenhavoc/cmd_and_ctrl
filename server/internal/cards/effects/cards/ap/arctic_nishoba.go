package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arctic Nishoba — Creature — Cat Warrior {5}{G}, 6/6 (Coldsnap, S82 OLD b558):
//
//	"Trample
//	 Cumulative upkeep {G} or {W}
//	 When this creature dies, you gain 2 life for each age counter on it."
//
// A hybrid cumulative upkeep (one {G/W} symbol per age counter) and a dies
// trigger that reads the Nishoba's last-known age counters (CR 603.10a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "66ae3349-ad37-47da-8ccc-3e1cdddb32a9",
		Name:            "Arctic Nishoba",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.CumulativeUpkeep("Arctic Nishoba — cumulative upkeep {G} or {W}", "{G/W}"),
			effects.WhenThisDies("Arctic Nishoba — you gain 2 life for each age counter on it",
				func(g *game.Game, item *game.StackItem) error {
					n := effects.AgeCountersAsItLeft(g, item)
					return effects.GainLife{Player: item.Controller, Amount: 2 * n}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
