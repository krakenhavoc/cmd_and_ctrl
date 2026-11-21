package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Woodland Champion — Creature — Elf Scout {1}{G}, 2/2:
//
//	"Whenever one or more tokens you control enter, put that many +1/+1
//	 counters on this creature."
//
// A batch trigger (CR 603.2c): the tokens that enter together are one
// trigger, and "that many" is how many of them entered in that batch,
// counted when the trigger resolves. The counters go on the Champion
// only if it is still the same object (CR 400.7). The Champion's own
// entering is not a token.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "675321d7-2cf0-4e0b-9517-d711b22865ab",
		Name:         "Woodland Champion",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventETB,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
					return ok && c.IsToken()
				},
				"Woodland Champion — that many +1/+1 counters on it (tokens entered)",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil || !effects.OnBattlefield(g, item.SourceCardID) || effects.SourceIsNewObject(g, item) {
						return nil
					}
					n := effects.BnsB42TokensInBatch(g, item.Controller, item.Trigger.Event.Batch)
					if n <= 0 {
						return nil
					}
					return effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: n}.Apply(effects.NewContext(g, item))
				})),
		},
	})
}
