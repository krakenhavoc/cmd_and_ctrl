package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Razor Pendulum — Artifact {4} (Mirage, #5633):
//
//	"At the beginning of each player's end step, if that player has 5 or less
//	 life, this artifact deals 2 damage to that player."
//
// An intervening "if" (CR 603.4): checked as the end step begins and again as the
// trigger resolves, against that player's life each time.
//
// No simplification.
func init() {
	atMost5 := func(g *game.Game, player uuid.UUID) bool {
		p := g.PlayerByIDForEffect(player)
		return p != nil && !p.Eliminated && p.Life <= 5
	}
	effects.Register(effects.Spec{
		OracleID:     "bf7c16db-01c9-41ed-bda6-c5326198ce77",
		Name:         "Razor Pendulum",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "that player has 5 or less life",
				Test: func(ctx *effects.Context) bool { return atMost5(ctx.Game, ctx.Trigger().Event.Actor) },
			}, effects.On(game.EventBeginEndStep, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return atMost5(g, ev.Actor)
			}, "Razor Pendulum — 2 damage to that player", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.DealDamage{Source: item.SourceCardID, Target: ctx.Trigger().Event.Actor, Amount: 2}.Apply(ctx)
			})),
		},
	})
}
