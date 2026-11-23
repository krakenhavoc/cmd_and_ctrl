package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reroute Systems — Instant {W} (Edge of Eternities, #4461):
//
//	"Choose one —
//	 • Target artifact or creature gains indestructible until end of turn.
//	 • Reroute Systems deals 2 damage to target tapped creature."
//
// Boros Charm's two grant and damage modes with a target each. The tapped
// clause is judged at announce and again at resolution (CR 608.2b), so a
// creature that untaps in response takes nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "26d2762b-4346-40d0-b93d-b0dbfeb6e2bc",
		Name:         "Reroute Systems",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.Mode("Target artifact or creature gains indestructible until end of turn.",
				effects.TargetPermanent("target artifact or creature", effects.Or(effects.Artifact(), effects.Creature()))),
				game.Purpose{Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "indestructible")), Answers: game.AnswerSurvive}),
			effects.ModeWithPurpose(effects.Mode("Reroute Systems deals 2 damage to target tapped creature.",
				effects.TargetCreature("target tapped creature", effects.TappedPermanent())),
				game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 2)), Answers: game.AnswerRemove}),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			switch {
			case ctx.HasMode(0):
				return effects.GrantKeywordUntilEOT{
					Target:   target,
					Keywords: []string{"indestructible"},
					Label:    "Reroute Systems — target artifact or creature gains indestructible",
				}.Apply(ctx)
			case ctx.HasMode(1):
				return effects.DealDamage{Source: ctx.Source(), Target: target, Amount: 2}.Apply(ctx)
			}
			return nil
		},
	})
}
