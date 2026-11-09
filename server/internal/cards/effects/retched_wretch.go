package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Retched Wretch — Creature — Goblin {2}{B}, 4/2 (Lorwyn Eclipsed, #4458):
//
//	"When this creature dies, if it had a -1/-1 counter on it, return it to
//	 the battlefield under its owner's control and it loses all
//	 abilities."
//
// An intervening "if" (CR 603.4) read from the Wretch as it last existed on
// the battlefield (CR 603.10a): it triggers only when it died with a -1/-1
// counter. A dead creature's counters cannot change, so the check as the
// trigger resolves would read the same record. The return is one card from the graveyard under its owner's
// control, and "it loses all abilities" is part of the entry
// (EntersAs.LosesAllAbilities, a layer 6 effect pinned to that object): it
// lasts as long as the object stays on the battlefield and does not expire
// at the cleanup step (the ruling), and the returned Wretch has no dies
// trigger, so it does not loop. A Wretch that left the graveyard in
// response is not returned.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0d3b6e27-2f8b-42e0-9490-7b134e84bb6b",
		Name:         "Retched Wretch",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return cardDied(ev, source) && retchedWretchHadACounter(g, ev.CardID)
			}, "Retched Wretch — return it to the battlefield, and it loses all abilities", retchedWretchReturn),
		},
	})
}

// retchedWretchHadACounter reads the counters the Wretch had as it last
// existed on the battlefield.
func retchedWretchHadACounter(g *game.Game, id uuid.UUID) bool {
	return g.LastKnownCountersForEffect(id)[game.CounterMinusOne] > 0
}

func retchedWretchReturn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return ReturnFromGraveyardTogether{
		Targets: []uuid.UUID{item.SourceCardID},
		EntersAs: &game.EntersAs{
			LosesAllAbilities: true,
			Label:             "Retched Wretch — loses all abilities",
		},
	}.Apply(ctx)
}
