package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b092_helpers.go — Zendikar Rising counter cards, S82 OLD batch 92 (#5508).

// S82OldB092FirstKickedSpellThisTurn passes on a kicked spell while its
// caster has cast no kicked spell yet this turn (Vine Gecko). The tally is
// bumped after a spell is priced (CR 601.2f), so none so far means this one
// is the first.
func S82OldB092FirstKickedSpellThisTurn() CostPredicate {
	return func(q game.CostQuery) bool {
		return q.Game != nil && game.KickedTimesPaid(q.Card, q.OptionalCosts) > 0 &&
			q.Game.CastTallyFor(q.Controller).Kicked == 0
	}
}

// S82OldB092PartyCounters is "put a +1/+1 counter on it for each creature
// in your party" (CR 700.8): the party is counted as the effect resolves.
func S82OldB092PartySize(ctx *Context) int {
	return FinB9PartySize(ctx.Game, ctx.Controller())
}
