package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Exemplar of Light — Creature — Angel {2}{W}{W}, 3/3 (EDHREC rank
// 1241):
//
//	"Flying
//	 Whenever you gain life, put a +1/+1 counter on this creature.
//	 Whenever you put one or more +1/+1 counters on this creature,
//	 draw a card. This ability triggers only once each turn."
//
// The lifegain half is Sanguine Bond's trigger with the drain swapped
// for a counter on itself (guarded on still being on the battlefield,
// since AddCounter does not gate on zone). The draw half is the first
// catalog trigger to watch EventCounterPlaced, and it needs three
// readings the event does not carry on its own:
//
//   - PLACEMENT, not removal: EventCounterPlaced fires for both and
//     carries only the new total, so b11CountersWerePlaced reads the
//     previous total back off the log.
//   - "YOU put": the placement names its placer on ev.Actor (ADR 0056
//     Decision 5; every placement since #3942), which
//     b11CounterPlacerOf reads. A counter placed by an opponent's
//     effect, or by an opponent's sandbox edit, does not draw.
//   - ONCE EACH TURN: the harvester emits EventTrigger, source and
//     label included, the moment it queues the ability, so
//     b11TriggeredThisTurn walks the log back to the current turn's
//     upkeep. A trigger that was countered still counts, as printed.
//
// "One or more" is one trigger per placement event, and a single
// AddCounter of N is one event, as printed; two separate placements
// in one turn are two events, of which only the first triggers.
//
// No simplification.
const b11ExemplarDrawLabel = "Exemplar of Light — draw a card"

func init() {
	effects.Register(effects.Spec{
		OracleID:        "9ad730b5-8950-4215-9319-387e2970dd22",
		Name:            "Exemplar of Light",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventChangeLife, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Target == source.Controller && ev.Amount > 0
			}, "Exemplar of Light — put a +1/+1 counter on it", effects.PutCounterOnSelf),
			effects.On(game.EventCounterPlaced, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B11CountersWerePlaced(ev, source.InstanceID, "+1/+1", g) &&
					effects.B11CounterPlacerOf(ev, g) == source.Controller &&
					!effects.B11TriggeredThisTurn(g, source.InstanceID, b11ExemplarDrawLabel)
			}, b11ExemplarDrawLabel, effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
