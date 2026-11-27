package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Intimidator Initiate — Creature — Goblin Shaman {R}, 1/1 (Shadowmoor, #5591):
//
//	"Whenever a player casts a red spell, you may pay {1}. If you do, target
//	 creature can't block this turn."
//
// Any player's red spell triggers it (read by the spell's current colours as
// it is cast). The target creature is chosen as the trigger goes on the
// stack (CR 603.3d); the payment is asked as it resolves, and only if it is
// made does the target, if still legal, stop being able to block until end
// of turn (CR 611.2c: it changes no characteristic).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "227f95ba-65b1-4399-8b12-39f8b1327d62",
		Name:         "Intimidator Initiate",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventCast, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor == uuid.Nil {
					return false
				}
				spell, ok := g.LookupCardForEffect(ev.CardID)
				return ok && spell.HasColor("R")
			}, "Intimidator Initiate — you may pay {1}: target creature can't block this turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}",
						Question: "Pay {1}: the target creature can't block this turn?",
						Then:     intimidatorInitiatePaid,
					}.Apply(effects.NewContext(g, item))
				}), effects.TargetCreature("target creature")),
		},
	})
}

// intimidatorInitiatePaid is the paid half, keyed: the target, if still
// legal, can't block this turn. An on-disk key: never renamed, never reused.
var intimidatorInitiatePaid = effects.PayUnlessBranches("pay-unless/intimidator-initiate-cant-block",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			return effects.RestrictUntilEOT{Target: t.ID, Restrictions: game.CantBlock, Label: "Intimidator Initiate — can't block"}.Apply(ctx)
		}
		return nil
	}, nil)
