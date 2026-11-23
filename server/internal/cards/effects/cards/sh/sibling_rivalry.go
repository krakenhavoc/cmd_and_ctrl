package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sibling Rivalry — Sorcery {3}{R} (The Brothers' War, #5482):
//
//	"Gain control of target artifact or creature until end of turn. Untap
//	 it. It gains haste until end of turn. Create a tapped Powerstone
//	 token."
//
// Act of Treason's three clauses on an artifact or creature, then the
// Powerstone. The Powerstone is made only when the target was still legal
// (otherwise the spell does not resolve, CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7fa1430f-cd2e-4071-a161-e32a61e33714",
		Name:         "Sibling Rivalry",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target artifact or creature", effects.Or(effects.Artifact(), effects.Creature())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			if err := (effects.GainControl{
				Target:   target,
				Duration: effects.DurationUntilEndOfTurn(ctx),
				Label:    "Sibling Rivalry — gain control until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.UntapTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.GrantKeywordUntilEOT{
				Target:   target,
				Keywords: []string{"haste"},
				Label:    "Sibling Rivalry — haste until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			return effects.CreateTokenAdvanced{Spec: effects.Token(effects.PowerstoneToken()).EntersTapped(), N: 1}.Apply(ctx)
		},
	})
}
