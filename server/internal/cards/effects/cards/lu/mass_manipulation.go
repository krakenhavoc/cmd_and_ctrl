package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mass Manipulation — Sorcery {X}{X}{U}{U}{U}{U} (Ravnica Allegiance):
//
//	"Gain control of X target creatures and/or planeswalkers."
//
// X is the number of targets (Back in Town's counted-by-X targets); control
// changes for good (CR 611.2a), each legal target taken at resolution.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "68e87559-7cae-4dcf-b7cd-a61f19bcd840",
		Name:         "Mass Manipulation",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Targets: effects.TargetsCountedByX(effects.TargetPermanent("X target creatures and/or planeswalkers",
			effects.Or(effects.Creature(), effects.Planeswalker()))),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.GainControl{Target: t.ID, Duration: game.IndefiniteDuration(),
					Label: "Mass Manipulation — gain control"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
