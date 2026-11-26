package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Azor's Elocutors — Creature — Human Advisor {3}{W/U}{W/U}, 3/5 (Return to
// Ravnica, #5566):
//
//	"At the beginning of your upkeep, put a filibuster counter on this creature.
//	 Then if this creature has five or more filibuster counters on it, you win the
//	 game.
//	 Whenever a source deals damage to you, remove a filibuster counter from this
//	 creature."
//
// The win is a "then if" inside the ability, not an intervening "if": it is read
// as the ability resolves, after the counter went on, and only while the creature
// is still on the battlefield (a creature that left in response has no counters to
// count). The second trigger fires once per damage event, whoever controls the
// source, and removes a counter only if there is one.
//
// No simplification.
func init() {
	const filibuster = "filibuster"
	effects.Register(effects.Spec{
		OracleID:     "45aeac84-4705-4179-be3f-b31bde844ea0",
		Name:         "Azor's Elocutors",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Azor's Elocutors — put a filibuster counter on it, then if it has five or more, you win the game", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if !effects.OnBattlefield(g, item.SourceCardID) || effects.SourceIsNewObject(g, item) {
					return nil
				}
				if err := (effects.AddCounter{Target: item.SourceCardID, Kind: filibuster, N: 1}).Apply(ctx); err != nil {
					return err
				}
				if c, ok := g.LookupCardForEffect(item.SourceCardID); ok && c.Counters[filibuster] >= 5 {
					return effects.WinTheGame{}.Apply(ctx)
				}
				return nil
			}),
			effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Amount > 0 && ev.Target == source.Controller
			}, "Azor's Elocutors — remove a filibuster counter from it", func(g *game.Game, item *game.StackItem) error {
				c, ok := g.LookupCardForEffect(item.SourceCardID)
				if !ok || !effects.OnBattlefield(g, item.SourceCardID) || effects.SourceIsNewObject(g, item) || c.Counters[filibuster] <= 0 {
					return nil
				}
				return effects.AddCounter{Target: item.SourceCardID, Kind: filibuster, N: -1}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
