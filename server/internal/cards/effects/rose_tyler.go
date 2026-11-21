package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rose Tyler — Legendary Creature — Human {1}{W}, 2/2 (Doctor Who, #5948):
//
//	"Rose Tyler gets +1/+1 for each time counter on it.
//	 Bad Wolf — Whenever Rose Tyler attacks, put a time counter on it for each
//	 suspended card you own and each other permanent you control with a time
//	 counter on it.
//	 Doctor's companion"
//
// The first ability is a layer-7c modification counting the time counters on
// Rose every pass. The attack trigger counts, as it resolves, the suspended
// cards the controller owns (in exile, with suspend and a time counter, CR
// 702.62b) and the other permanents they control with a time counter, and puts
// that many time counters on Rose. Rose herself is not counted, so it does
// not double each attack. Doctor's companion is a deck-construction rule;
// "Bad Wolf" is an ability word (CR 207.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2f56de72-80d7-48c6-9b57-84fbd252c699",
		Name:         "Rose Tyler",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Rose Tyler gets +1/+1 for each time counter on it.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
				n := source.Counters[game.CounterTime]
				c.Power += n
				c.Toughness += n
			},
		}},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Rose Tyler — put a time counter on it for each suspended card you own and each other permanent you control with a time counter on it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					n := 0
					for _, id := range timeTravelObjects(g, item.Controller) {
						if id != uuid.Nil && id != item.SourceCardID {
							n++
						}
					}
					if n == 0 || !onBattlefield(g, item.SourceCardID) {
						return nil
					}
					return AddCounter{Target: item.SourceCardID, Kind: game.CounterTime, N: n}.Apply(ctx)
				}),
		},
	})
}
