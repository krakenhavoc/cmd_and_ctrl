package cardshu

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Iridescent Hornbeetle — {4}{G} Creature — Insect 3/4 (Jumpstart 2022, #5948):
//
//	"At the beginning of your end step, create a 1/1 green Insect creature
//	 token for each +1/+1 counter you've put on creatures under your control
//	 this turn."
//
// Counted off the turn's event log as the trigger resolves: every +1/+1
// counter you put (the placing player is on the event), counting the
// counters, not the placements, on a permanent that is, or as it last
// existed was, a creature under your control. Counters a creature entered
// with are put by you (CR 122.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3270265f-fa9f-4ef0-a1a3-a68fc2349bf1",
		Name:         "Iridescent Hornbeetle",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Iridescent Hornbeetle — create a 1/1 green Insect creature token for each +1/+1 counter you've put on creatures under your control this turn",
				func(g *game.Game, item *game.StackItem) error {
					n := hornbeetleCountersPutThisTurn(g, item.Controller)
					if n <= 0 {
						return nil
					}
					return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("1/1 green Insect"), N: n}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// hornbeetleCountersPutThisTurn counts the +1/+1 counters `you` put this
// turn on creatures under your control.
func hornbeetleCountersPutThisTurn(g *game.Game, you uuid.UUID) int {
	n := 0
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind != game.EventCounterPlaced || ev.Label != game.CounterPlusOne || ev.Target == uuid.Nil || ev.Actor != you {
			continue
		}
		put := ev.Amount - g.CounterCountBeforeForEffect(ev)
		if put <= 0 {
			continue
		}
		ref, ok := g.PermanentRefForEffect(ev.Target)
		if !ok {
			continue
		}
		if info, ok := g.PermanentForEffect(ref); ok && info.Controller == you && hornbeetleIsCreature(info.Characteristic.Types) {
			n += put
		}
	}
	return n
}

func hornbeetleIsCreature(types []string) bool {
	for _, t := range types {
		if strings.EqualFold(t, "creature") {
			return true
		}
	}
	return false
}
