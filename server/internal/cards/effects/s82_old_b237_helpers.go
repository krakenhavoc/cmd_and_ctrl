package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// ExileEachLegalCardTarget is "exile <the targets>": every card target still
// legal as the ability resolves is exiled, in the order announced (CR 608.2b).
// Dust, Desecrate Reality.
func ExileEachLegalCardTarget(ctx *Context) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
