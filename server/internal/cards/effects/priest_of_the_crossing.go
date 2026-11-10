package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Priest of the Crossing — Creature — Zombie Bird Cleric {3}{W}, 3/3:
//
//	"Flying
//	 At the beginning of each end step, put X +1/+1 counters on each
//	 creature you control, where X is the number of creatures that died
//	 under your control this turn."
//
// "Each end step" is any player's. X is the controller's per-turn tally of
// creatures that died while they controlled them (game.PlayerTurnTally.
// CreaturesDied, tokens included), read once as the trigger resolves; each
// creature then gets X counters as its own placement, so a counter
// replacement applies to each.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "524e5098-9c6d-43ad-80cb-52fd6d14ef53",
		Name:            "Priest of the Crossing",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return b15EndStepBegan(ev)
			}, "Priest of the Crossing — put X +1/+1 counters on each creature you control, where X is the number of creatures that died under your control this turn",
				func(g *game.Game, item *game.StackItem) error {
					x := g.TurnTallyFor(item.Controller).CreaturesDied
					if x <= 0 {
						return nil
					}
					ctx := NewContext(g, item)
					var ids []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == item.Controller && c.IsCreature() {
							ids = append(ids, c.InstanceID)
						}
					}
					for _, id := range ids {
						if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: x}).Apply(ctx.asGroupMember()); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}
