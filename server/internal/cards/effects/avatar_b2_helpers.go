package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// avatar_b2_helpers.go — Avatar batch 2 (#4459).

// airbendFirstLegalTarget is "airbend up to one target creature" as a
// trigger body (Glider Staff): the first target still legal on
// resolution is airbent (CR 608.2b); none, and nothing happens.
// Unlike AirbendOtherTarget it has no "other" clause to enforce.
func airbendFirstLegalTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return Airbend{Target: t.ID}.Apply(ctx)
		}
	}
	return nil
}
