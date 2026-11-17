package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mat_b04_helpers.go — shared pieces of the S82 March of the Machine:
// The Aftermath batch 4 cards (#5476).

// ManaValueLessThanPower is "with mana value less than this creature's
// power" (Narset, Enlightened Exile): the strict sibling of
// ManaValueNoGreaterThanPower, judged as the target is chosen and again as
// the ability resolves (CR 608.2b), against last-known power (CR 608.2h).
func ManaValueLessThanPower() game.SourceComparison {
	return game.SourceComparison{Stat: game.SourceManaValueVsPower, Cmp: game.CmpLess}
}

// matB04ExileTargetsAndCopy exiles the still-legal card targets of the
// resolving ability and hands what really reached exile to the given tail,
// which copies it and offers the cast (CR 707.12).
func matB04ExileTargetsAndCopy(tail game.TailRef) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		var ids []uuid.UUID
		for _, t := range ctx.LegalTargets() {
			if t.Kind == game.TargetCard {
				ids = append(ids, t.ID)
			}
		}
		if len(ids) == 0 {
			return nil
		}
		return g.ExileCardsKeyedForEffect(ids, NewContext(g, item).ItemTail(tail))
	}
}
