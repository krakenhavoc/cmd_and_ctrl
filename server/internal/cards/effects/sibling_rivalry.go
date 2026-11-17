package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "7fa1430f-cd2e-4071-a161-e32a61e33714",
		Name:         "Sibling Rivalry",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target artifact or creature", Or(Artifact(), Creature())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			if err := (GainControl{
				Target:   target,
				Duration: DurationUntilEndOfTurn(ctx),
				Label:    "Sibling Rivalry — gain control until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			if err := (UntapTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			if err := (GrantKeywordUntilEOT{
				Target:   target,
				Keywords: []string{"haste"},
				Label:    "Sibling Rivalry — haste until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			return CreateTokenAdvanced{Spec: Token(PowerstoneToken()).EntersTapped(), N: 1}.Apply(ctx)
		},
	})
}
