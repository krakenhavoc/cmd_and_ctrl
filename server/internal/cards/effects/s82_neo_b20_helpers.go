package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_neo_b20_helpers.go — shared vocabulary for the S82 NEO batch 20 cards.

// neoB20BounceTargetedPermanents is "return <the targets> to their owners'
// hands": every card target still legal as the ability resolves goes back, in
// the order they were announced (Kairi, the Swirling Sky's first bullet).
func NeoB20BounceTargetedPermanents(ctx *Context) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (BounceToHand{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
