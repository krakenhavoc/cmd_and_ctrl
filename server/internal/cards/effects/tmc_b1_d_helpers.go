package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tmc_b1_d_helpers.go — shared pieces for TMNT Commander batch 1,
// slice d (#5011): "for each opponent, create a token that's tapped and
// attacking that player". Squad (CR 702.157) lives in tmc_b4_helpers.go.

// tokenTappedAndAttackingEachOpponent is "for each opponent, create
// <tmpl> that's tapped and attacking that player": one token per
// opponent still in the game as the ability resolves, whether or not
// that player was attacked. The tokens were never declared as attackers,
// so nothing that triggers on a creature attacking sees them (CR 508.4).
func tokenTappedAndAttackingEachOpponent(ctx *Context, tmpl func() game.Card) error {
	for _, opp := range ctx.Opponents() {
		t := tmpl()
		t.Tapped = true
		t.AttackingTarget = opp
		if err := (CreateToken{Controller: ctx.Controller(), Template: t, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
