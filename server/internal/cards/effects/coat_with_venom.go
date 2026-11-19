package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Coat with Venom — Instant {B} (Foundations Jumpstart, #5929):
//
//	"Target creature gets +1/+2 and gains deathtouch until end of turn."
//
// The boost and the keyword are both turn-scoped effects on the target as
// it resolves (CR 611.2c); a target that is no longer legal fizzles the
// whole spell (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cbc808dd-8d50-4f42-b0f1-d4e45d0462a2",
		Name:         "Coat with Venom",
		Purpose:      ForTargets(BenefitUntilEOT(0, 1, 2, "deathtouch")),
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			if err := (BoostUntilEOT{Target: target, Power: 1, Toughness: 2, Label: "Coat with Venom — +1/+2"}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{Target: target, Keywords: []string{"deathtouch"}, Label: "Coat with Venom — deathtouch"}.Apply(ctx)
		},
	})
}
