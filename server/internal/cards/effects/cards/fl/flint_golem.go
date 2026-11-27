package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Flint Golem — Artifact Creature — Golem {4}, 2/3 (Nemesis, S82 OLD batch
// 680, #5618):
//
//	"Whenever this creature becomes blocked, defending player mills three
//	 cards."
//
// One trigger per blocked attacker however many creatures block it (CR
// 509.3c, CR 506.4). The defending player is the becomes-blocked event's
// actor, the controller of its blockers, and is not a target; a player who
// has left the game mills nothing.
//
// No simplification.
func init() {
	const label = "Flint Golem — defending player mills three cards"
	effects.Register(effects.Spec{
		OracleID:     "0993c5f1-1e8f-4871-9904-6c5a00405f36",
		Name:         "Flint Golem",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesBlocked, effects.Self, label, func(g *game.Game, item *game.StackItem) error {
				if item.Trigger == nil {
					return nil
				}
				defender := item.Trigger.Event.Actor
				if g.PlayerByIDForEffect(defender) == nil {
					return nil
				}
				return effects.MillCards{Player: defender, N: 3}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
