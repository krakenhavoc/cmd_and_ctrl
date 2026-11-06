package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Meltstrider Eulogist — Creature — Insect Soldier {2}{G}, 3/3 (Edge of
// Eternities, #4461):
//
//	"Whenever a creature you control with a +1/+1 counter on it dies,
//	 draw a card."
//
// Zulaport Cutthroat's death watch (ACreatureYouControlDied, the Eulogist
// itself included) narrowed by Necroskitter's read of the counters the
// creature had as it last existed on the battlefield
// (LastKnownCountersForEffect): a leaves-the-battlefield trigger looks
// back in time (CR 603.10a), so the counters count although the card in
// the graveyard has none. A creature with only other kinds of counter
// (a -1/-1 counter, a shield counter) does not draw.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "046b60da-0a14-40b0-a36c-5328f6b8972b",
		Name:         "Meltstrider Eulogist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(On(game.EventLTB, func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
				return ACreatureYouControlDied(ev, source, ch, g) && g.LastKnownCountersForEffect(ev.CardID)[game.CounterPlusOne] > 0
			}, "Meltstrider Eulogist — draw a card", Do(DrawCards{N: 1})), game.Purpose{DeathPayoff: true}),
		},
	})
}
