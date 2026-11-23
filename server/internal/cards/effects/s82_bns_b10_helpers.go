package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// detainEachLegalTarget is "detain <target creatures>": every target still
// legal as the ability resolves is detained (CR 701.35a), the set pinned
// now, until the controller's next turn (Detain).
func DetainEachLegalTarget(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			id := t.ID
			if err := (Detain{
				Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.InstanceID == id },
				Label: label,
			}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	}
}

// bnsB10HandAtMostOne is "you have one or fewer cards in hand", read for the
// source's controller on every recompute (CR 611.3a).
func BnsB10HandAtMostOne(g *game.Game, you uuid.UUID) bool {
	p := g.PlayerByIDForEffect(you)
	return p != nil && p.Hand != nil && p.Hand.Size() <= 1
}

// bnsB10DiscardHandThenInvestigate is Carnage Interpreter's enter body:
// discard your hand (even an empty one), then investigate four times.
func BnsB10DiscardHandThenInvestigate(g *game.Game, item *game.StackItem) error {
	if _, err := DiscardWholeHand(g, item.Controller); err != nil {
		return err
	}
	return Investigate{N: 4}.Apply(NewContext(g, item))
}
