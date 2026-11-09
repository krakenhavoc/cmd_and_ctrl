package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "26d2762b-4346-40d0-b93d-b0dbfeb6e2bc",
		Name:         "Reroute Systems",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeWithPurpose(Mode("Target artifact or creature gains indestructible until end of turn.",
				TargetPermanent("target artifact or creature", Or(Artifact(), Creature()))),
				game.Purpose{Targets: game.ForTargets(BenefitUntilEOT(0, 0, 0, "indestructible")), Answers: game.AnswerSurvive}),
			ModeWithPurpose(Mode("Reroute Systems deals 2 damage to target tapped creature.",
				TargetCreature("target tapped creature", tappedPermanent())),
				game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 2)), Answers: game.AnswerRemove}),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			switch {
			case ctx.HasMode(0):
				return GrantKeywordUntilEOT{
					Target:   target,
					Keywords: []string{"indestructible"},
					Label:    "Reroute Systems — target artifact or creature gains indestructible",
				}.Apply(ctx)
			case ctx.HasMode(1):
				return DealDamage{Source: ctx.Source(), Target: target, Amount: 2}.Apply(ctx)
			}
			return nil
		},
	})
}
