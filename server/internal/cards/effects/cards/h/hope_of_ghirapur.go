package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hope of Ghirapur — Legendary Artifact Creature — Thopter {1}, 1/1 (Aether
// Revolt, S82 older sets, #5536):
//
//	"Flying
//	 Sacrifice Hope of Ghirapur: Until your next turn, target player who was
//	 dealt combat damage by Hope of Ghirapur this turn can't cast noncreature
//	 spells."
//
// The target clause is judged against the Thopter as the object it was (a
// last-known record once it has been sacrificed as the cost, CR 608.2h): the
// player must have been dealt combat damage by that object this turn, at
// announcement and again as the ability resolves (CR 608.2b)
// (game.TargetSpec.DealtCombatDamageBySource). The ban is a granted cast ban
// on that player that lasts until the activator's next turn begins.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1a20e805-c80b-4710-917b-706d29ec395e",
		Name:            "Hope of Ghirapur",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice Hope of Ghirapur: Until your next turn, target player who was dealt combat damage by Hope of Ghirapur this turn can't cast noncreature spells.",
			Purpose: game.Purpose{Answers: game.AnswerRestrict},
			Cost:    effects.SacrificeThis(),
			Targets: effects.DealtCombatDamageBySource(effects.TargetPlayer("target player who was dealt combat damage by Hope of Ghirapur this turn")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetPlayer {
						continue
					}
					g.GrantCastBanForEffect(t.ID, game.CastBanRule{
						Kind:   game.CastBanOutright,
						Filter: game.PermissionFilter{NoncreatureOnly: true},
					}, "Hope of Ghirapur — can't cast noncreature spells", ctx.Source(),
						effects.DurationUntilYourNextTurn(ctx, ctx.Controller()))
				}
				return nil
			},
		}},
	})
}
