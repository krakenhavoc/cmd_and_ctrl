package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spelltithe Enforcer — Creature — Elephant Wizard {3}{W}{W}, 3/3
// (Guildpact):
//
//	"Whenever an opponent casts a spell, that player sacrifices a
//	 permanent of their choice unless they pay {1}."
//
// Rhystic Study's trigger with the caster as the payer: the pay-unless is
// asked of the opponent who cast the spell, and on a decline that player,
// not the Enforcer's controller, sacrifices a permanent of their choice.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0da4988f-47ca-4786-ac33-edbc9cfd6785",
		Name:         "Spelltithe Enforcer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.ByAnOpponent,
				"Spelltithe Enforcer — that player sacrifices a permanent unless they pay {1}",
				func(g *game.Game, item *game.StackItem) error {
					return effects.PayUnless{
						Chooser:  item.Trigger.Event.Actor,
						Cost:     "{1}",
						Question: "Spelltithe Enforcer — pay {1}? (Otherwise sacrifice a permanent)",
						Else:     spelltitheEnforcerDecline,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// spelltitheEnforcerDecline is the sacrifice, keyed: a table waiting on the
// payer is a restore point.
var spelltitheEnforcerDecline = effects.PayUnlessElse("pay-unless/spelltithe-enforcer-sacrifice", func(ctx *effects.Context, r game.PayUnlessAnswer) error {
	ctx.Game.PlayerSacrificesForEffect(ctx.Source(), r.Chooser, effects.SacrificeSpec("a permanent", effects.Permanent()),
		"Spelltithe Enforcer — sacrifice a permanent")
	return nil
})
