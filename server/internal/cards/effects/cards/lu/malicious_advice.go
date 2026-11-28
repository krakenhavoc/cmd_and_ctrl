package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Malicious Advice — Instant {X}{U}{B} (Planeshift, S82 OLD b642, #5615):
//
//	"Tap X target artifacts, creatures, and/or lands. You lose X life."
//
// The targets are Icy Blast's "X target" clause (the count is the announced X,
// CR 107.3a) over artifacts, creatures and lands. Each legal target is tapped as
// the spell resolves (CR 608.2b); the life loss is X whatever happened to the
// targets, so long as one is still legal (a spell whose targets are all illegal
// does not resolve, CR 608.2b). At X = 0 there are no targets and it resolves,
// losing nothing.
//
// No simplification.
func init() {
	targets := effects.TargetPermanent("X target artifacts, creatures, and/or lands", effects.Or(effects.Artifact(), effects.Creature(), effects.Land()))
	targets.CountFromX = true
	effects.Register(effects.Spec{
		OracleID:     "14c1adf8-950b-4f64-9c84-956a623448f4",
		Name:         "Malicious Advice",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Targets:      targets,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.TapTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -ctx.X())
		},
	})
}
