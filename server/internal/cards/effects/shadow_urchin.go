package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shadow Urchin — Creature — Ouphe {2}{B/R} (Lorwyn Eclipsed, #4458):
//
//	"Whenever this creature attacks, blight 1.
//	 Whenever a creature you control with one or more counters on it dies,
//	 exile that many cards from the top of your library. Until your next
//	 end step, you may play those cards."
//
// "That many" is every counter of any kind the creature had as it died
// (last-known information, CR 603.10a; the 2026-rulings: +1/+1 and -1/-1
// counters both count). The cards are played, so a land spends a land drop,
// and all costs and timing apply (the ruling).
func init() {
	Register(Spec{
		OracleID:     "86bd332f-ba61-49a4-ad31-de4d7e3638c6",
		Name:         "Shadow Urchin",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Shadow Urchin — blight 1", func(g *game.Game, item *game.StackItem) error {
				return Blight{N: 1}.Apply(NewContext(g, item))
			}),
			TriggerWithPurpose(On(game.EventLTB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return ACreatureYouControlDied(ev, source, lki, g) && shadowUrchinCounters(g, ev.CardID) > 0
			}, "Shadow Urchin — exile that many cards from the top of your library; until your next end step, you may play them",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					// Last-known information as it died (CR 603.10a): the record
					// outlives the dying snapshot the trigger's condition reads.
					info, ok := ctx.TriggeringPermanent()
					if !ok {
						return nil
					}
					n := 0
					for _, c := range info.Counters {
						n += c
					}
					if n <= 0 {
						return nil
					}
					d := g.UntilYourNextEndStepDuration(item.Controller)
					return ExileTopWithPermission{From: item.Controller, GrantTo: item.Controller, N: n, Window: &d}.Apply(ctx)
				}), game.Purpose{DeathPayoff: true}),
		},
	})
}

// shadowUrchinCounters is how many counters, of any kind, the creature had
// on the battlefield as it died.
func shadowUrchinCounters(g *game.Game, id uuid.UUID) int {
	n := 0
	for _, c := range g.LastKnownCountersForEffect(id) {
		n += c
	}
	return n
}
