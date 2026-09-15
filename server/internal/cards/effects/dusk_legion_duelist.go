package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dusk Legion Duelist — Creature — Vampire Soldier {1}{W}, 2/2
// (EDHREC rank 2750):
//
//	"Vigilance
//	 Whenever one or more +1/+1 counters are put on this creature,
//	 draw a card. This ability triggers only once each turn."
//
// The counters deck's two-drop cantrip. Exemplar of Light's draw
// trigger without the "you put" clause — anyone's counters count:
//
//   - PLACEMENT, not removal: EventCounterPlaced fires for both and
//     carries only the new total, so b11CountersWerePlaced reads the
//     previous total back off the log.
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
const b26DuskLegionDuelistDrawLabel = "Dusk Legion Duelist — draw a card"

func init() {
	Register(Spec{
		OracleID:        "30367a59-a812-41d4-a451-80d9071dd4a4",
		Name:            "Dusk Legion Duelist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			On(game.EventCounterPlaced, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b11CountersWerePlaced(ev, source.InstanceID, "+1/+1", g) &&
					!b11TriggeredThisTurn(g, source.InstanceID, b26DuskLegionDuelistDrawLabel)
			}, b26DuskLegionDuelistDrawLabel, Do(DrawCards{N: 1})),
		},
	})
}
