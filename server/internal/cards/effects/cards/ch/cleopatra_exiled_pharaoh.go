package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cleopatra, Exiled Pharaoh — Legendary Creature — Human Noble {2}{B}{G},
// 2/4 (Assassin's Creed, #5948):
//
//	"Allies — At the beginning of your end step, put a +1/+1 counter on
//	 each of up to two other target legendary creatures.
//	 Betrayal — Whenever a legendary creature with counters on it dies,
//	 draw a card for each counter on it. You lose 2 life."
//
// Allies is Aerie Auxiliary's support shape on an end-step trigger, with a
// legendary filter and "other" (the source itself is excluded); a target
// gone by resolution takes nothing (CR 608.2b).
//
// Betrayal is any legendary creature's death, whoever controlled it,
// judged as it last existed (CR 603.10a: Event.WasSupertype, the exit's
// last-known counters). The cards drawn are the total number of counters of
// every kind it had, read off the event's last-known snapshot as the trigger
// resolves; the life loss follows the draw. A legendary creature with no
// counters does not trigger it.
//
// No simplification.
func init() {
	const betrayal = "Cleopatra, Exiled Pharaoh — draw a card for each counter on it, then you lose 2 life"
	effects.Register(effects.Spec{
		OracleID:     "e7a80464-1522-48d7-bf82-f9894f5c3ae0",
		Name:         "Cleopatra, Exiled Pharaoh",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.AtYourEndStep("Cleopatra, Exiled Pharaoh — put a +1/+1 counter on each of up to two other target legendary creatures",
				effects.PutPlusOneCounterOnEachLegalTarget),
				effects.Another(effects.TargetCreature("up to two other target legendary creatures", effects.Legendary()).WithCount(0, 2))),
			{
				Watches: []game.EventKind{game.EventLTB},
				AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
					if _, ok := effects.DiedCreature(ev, g); !ok {
						return false
					}
					if was, known := ev.WasSupertype("legendary"); !known || !was {
						return false
					}
					return cleopatraCounters(g, ev) > 0
				},
				Key: betrayal,
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					// The counters are the snapshot the event carries of the
					// creature as it last existed (CR 603.10a).
					n := 0
					if obj := ctx.Trigger().Object; obj != nil {
						for _, c := range obj.Counters {
							n += c
						}
					}
					if err := (effects.DrawCards{N: n}).Apply(ctx); err != nil {
						return err
					}
					return effects.YouLoseLife(2)(ctx)
				},
			},
		},
	})
}

// cleopatraCounters is how many counters, of any kind, the creature the
// event names had as it last existed on the battlefield.
func cleopatraCounters(g *game.Game, ev game.Event) int {
	n := 0
	for _, c := range g.LastKnownCountersForEffect(ev.CardID) {
		n += c
	}
	return n
}
