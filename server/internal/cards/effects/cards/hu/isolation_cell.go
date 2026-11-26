package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Isolation Cell — Artifact {4}:
//
//	"Whenever an opponent casts a creature spell, that player loses 2 life
//	 unless they pay {2}."
//
// The caster is asked, and the life loss is theirs.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "56c558f3-535d-49aa-a2de-54e0bb598483",
		Name:         "Isolation Cell",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.AnOpponentCast(effects.Creature()),
				"Isolation Cell — that player loses 2 life unless they pay {2}",
				func(g *game.Game, item *game.StackItem) error {
					return effects.PayUnless{
						Chooser:  item.Trigger.Event.Actor,
						Cost:     "{2}",
						Question: "Isolation Cell — pay {2} or lose 2 life?",
						Else:     isolationCellDecline,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

var isolationCellDecline = effects.PayUnlessElse("pay-unless/s82-old-b459-isolation-cell-life", func(ctx *effects.Context, r game.PayUnlessAnswer) error {
	return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), r.Chooser, -2)
})
