package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kid Loki — Legendary Creature — God Hero Villain {U}, 1/1 (Marvel Super
// Heroes, #4455):
//
//	"Each creature you control that you've put one or more +1/+1
//	 counters on this turn has hexproof.
//	 Whenever you draw your second card each turn, put a +1/+1 counter
//	 on Kid Loki."
//
// The hexproof is a layer-6 static judged per creature against this turn's
// event log (Game.EventsThisTurn): a +1/+1 placement by the source's
// controller on that very object, counted from the object's last arrival
// on the battlefield (CR 400.7, so a creature that left and came back is
// a new one). It ends with the turn because the log does. The second-card
// trigger is Atlantean Cavalry's, and the counter it puts on Kid Loki is
// one he is granted hexproof by from then on.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "19ef0329-08de-4835-8079-2e791e19f3a6",
		Name:         "Kid Loki",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label: "Each creature you control that you've put one or more +1/+1 counters on this turn has hexproof.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller &&
					kidLokiPutPlusOneOn(g, source.Controller, target.InstanceID)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				if !slices.Contains(c.Abilities, "hexproof") {
					c.Abilities = append(c.Abilities, "hexproof")
				}
			},
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Kid Loki — put a +1/+1 counter on Kid Loki",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}

// kidLokiPutPlusOneOn reports whether `you` put a +1/+1 counter on the
// permanent `id` this turn, since its last arrival on the battlefield.
// Caller must hold g.mu (a static's AppliesTo does).
func kidLokiPutPlusOneOn(g *game.Game, you, id uuid.UUID) bool {
	events := g.EventsThisTurn()
	for i := len(events) - 1; i >= 0; i-- {
		ev := events[i]
		if ev.Kind == game.EventLTB && ev.CardID == id {
			return false
		}
		if ev.Kind == game.EventCounterPlaced && ev.Target == id && ev.Label == game.CounterPlusOne && ev.Actor == you &&
			counterPlacementDelta(ev, g) > 0 {
			return true
		}
	}
	return false
}
