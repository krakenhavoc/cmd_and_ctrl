package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Taurean Mauler — Creature — Shapeshifter {2}{R}, 2/2:
//
//	"Changeling (This card is every creature type.)
//	 Whenever an opponent casts a spell, you may put a +1/+1 counter
//	 on this creature."
//
// Changeling needs no catalog entry (deck import stamps it, S26); the
// catalog entry exists only for the "you may" counter trigger. "An
// opponent casts A spell" — any spell, no type filter — needs no
// event-count tally, just an ordinary EventCast with the actor being
// anyone but the controller, and the OptionalPrompt gate makes the
// "may" a real question rather than an assumed yes.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fd5ee26c-cc54-4dc0-b611-5cda14354a5e",
		Name:         "Taurean Mauler",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor != uuid.Nil && ev.Actor != source.Controller
			},
			Key:    "Taurean Mauler — put a +1/+1 counter on this creature",
			Effect: taureanMaulerAddCounter,
			OptionalPrompt: &game.TriggerOptionalPrompt{
				Question: "Taurean Mauler — put a +1/+1 counter on it?",
			},
		}},
	})
}

func taureanMaulerAddCounter(g *game.Game, item *game.StackItem) error {
	return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(NewContext(g, item))
}
